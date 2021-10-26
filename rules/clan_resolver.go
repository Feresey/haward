package rules

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/ratelimit"
)

type Clan struct {
	Name, Tag string
}

type PlayerClanResolver struct {
	// map[player_name]clan
	cache map[string]Clan

	rl  ratelimit.Limiter
	cli *http.Client
}

func NewPlayerResolver() *PlayerClanResolver {
	return &PlayerClanResolver{
		cache: make(map[string]Clan),
		rl:    ratelimit.New(30), // 0.3 second
		cli:   http.DefaultClient,
	}
}

func (p *PlayerClanResolver) GetPlayerClan(nickname string) (*Clan, error) {
	cached, ok := p.cache[nickname]
	if ok {
		return &cached, nil
	}
	clan, err := p.getFromAPI(nickname)
	if err != nil {
		return nil, fmt.Errorf("get player clan from api: %s, %w", nickname, err)
	}

	p.cache[nickname] = *clan
	return clan, nil
}

func (p *PlayerClanResolver) getFromAPI(nickname string) (*Clan, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://gmt.star-conflict.com/pubapi/v1/userinfo.php?nickname=%s", nickname),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	p.rl.Take()

	response, err := p.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer response.Body.Close()

	var data struct {
		Data struct {
			Clan struct {
				Name string `json:"name,omitempty"`
				Tag  string `json:"tag,omitempty"`
			} `json:"clan,omitempty"`
		} `json:"data"`
	}

	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &Clan{Name: data.Data.Clan.Name, Tag: data.Data.Clan.Tag}, nil
}

func (p *PlayerClanResolver) AddOldNickname(oldNickname, newNickname string) error {
	clanCached, ok := p.cache[newNickname]
	if ok {
		p.cache[oldNickname] = clanCached
	}

	clan, err := p.GetPlayerClan(newNickname)
	if err != nil {
		return err
	}

	p.cache[oldNickname] = *clan
	return nil
}
