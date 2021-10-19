package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/Feresey/haward/parse"
)

type Rules struct {
	// награды за конкретных пилотов
	awards map[string]int
	// повелители бури
	punishments map[string]int
	// теги кланов, за которыми охота
	clanTags map[string]int
	// полные названия кланов, за которыми охота
	clanNames map[string]int
}

func (r *Rules) CheckAward(nickname, clanTag, clanFullName string) (award int, ok bool) {
	award, ok = r.awards[nickname]
	if ok {
		return
	}
	award, ok = r.punishments[nickname]
	if ok {
		return
	}
	award, ok = r.clanTags[clanTag]
	if ok {
		return
	}
	award, ok = r.clanNames[clanFullName]
	if ok {
		return
	}
	return 0, false
}

type Parser struct {
	yourNickname string
	rules        Rules

	resolver   *PlayerClanResolver
	levelIter  *parse.GameLogIter
	gameReader io.Reader

	lastLevel bool
}

func NewParser(
	yourNickname string,
	combat, game io.Reader,
) *Parser {
	return &Parser{
		yourNickname: yourNickname,
		levelIter:    parse.NewGameLogIter(yourNickname, game),
		resolver:     NewPlayerResolver(),
		gameReader:   game,
	}
}

func (p *Parser) Parse() error {
	for !p.lastLevel {
		err := p.ParseLogLevel()
		if err != nil {
			return err
		}

	}
	return nil
}

func (p *Parser) ParseLogLevel() error {
	lvl, err := p.levelIter.ScanNextLevel()
	if err != nil {
		if errors.Is(err, io.EOF) {
			p.lastLevel = true
		}
		return fmt.Errorf("parse log level: %w", err)
	}

	awards, err := p.getEnemiesAwards(lvl)
	if err != nil {
		return fmt.Errorf("get awards: %w", err)
	}

	// TODO
	_ = awards

	return nil
}

func (p *Parser) getEnemiesAwards(lvl *parse.GameLogLevel) (map[string]int, error) {
	awards := make(map[string]int)
	for _, enemy := range lvl.GetEnemies() {
		award, ok := p.rules.CheckAward(enemy.Name, enemy.ClanTag, "")
		if ok {
			awards[enemy.Name] = award
			continue
		}

		clan, err := p.resolver.GetPlayerClanName(enemy.Name)
		if err != nil {
			return nil, err
		}

		award, ok = p.rules.CheckAward(enemy.Name, enemy.ClanTag, clan)
		if ok {
			awards[enemy.Name] = award
			continue
		}
	}
	return awards, nil
}

type PlayerClanResolver struct {
	// map[player_name]clan_name
	cache map[string]string
}

func NewPlayerResolver() *PlayerClanResolver {
	return &PlayerClanResolver{
		cache: make(map[string]string),
	}
}

func (p *PlayerClanResolver) GetPlayerClanName(nickname string) (string, error) {
	cached, ok := p.cache[nickname]
	if ok {
		return cached, nil
	}

	// TODO curl http://gmt.star-conflict.com/pubapi/v1/userinfo.php\?nickname\=AlaStoR
	p.cache[nickname] = ""

	return "", nil
}
