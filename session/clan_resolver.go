package session

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/ratelimit"
)

type PlayerClanResolver struct {
	// map[player_name]clan_name
	cache map[string]string

	rl  ratelimit.Limiter
	cli *http.Client
}

func NewPlayerResolver() *PlayerClanResolver {
	return &PlayerClanResolver{
		cache: make(map[string]string),
		rl:    ratelimit.New(30), // 0.3 second
		cli:   http.DefaultClient,
	}
}

func (p *PlayerClanResolver) GetPlayerClanName(nickname string) (string, error) {
	cached, ok := p.cache[nickname]
	if ok {
		return cached, nil
	}

	name, err := p.getFromAPI(nickname)
	if err != nil {
		return "", fmt.Errorf("get player clan from api: %s, %w", nickname, err)
	}

	p.cache[nickname] = *name

	return *name, nil
}

func (p *PlayerClanResolver) getFromAPI(nickname string) (*string, error) {
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

	return &data.Data.Clan.Name, nil
}
