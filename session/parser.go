package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/Feresey/haward/parse"
)

// Rules описывает правила ивента - награды за головы и штрафы.
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

// Parser это сущность которая обрабатывает логи одной сессии
type Parser struct {
	yourNickname string
	rules        Rules

	resolver       *PlayerClanResolver
	levelIter      *parse.GameLogIter
	gameLogScanner *bufio.Scanner

	lastLevel bool
}

func NewParser(
	yourNickname string,
	combat, game io.Reader,
	rules Rules,
) *Parser {
	return &Parser{
		yourNickname:   yourNickname,
		rules:          rules,
		lastLevel:      false,
		levelIter:      parse.NewGameLogIter(yourNickname, game),
		resolver:       NewPlayerResolver(),
		gameLogScanner: bufio.NewScanner(combat),
	}
}

func (p *Parser) Parse(results chan<- []parse.DeathRecord) error {
	for !p.lastLevel {
		scores, err := p.ParseLogLevel()
		if err != nil {
			return err
		}

		results <- scores
	}
	return nil
}

// ParseLogLevel парсит один уровень (одну игру по идее)
func (p *Parser) ParseLogLevel() ([]parse.DeathRecord, error) {
	lvl, err := p.levelIter.ScanNextLevel()
	if err != nil {
		if errors.Is(err, io.EOF) {
			p.lastLevel = true
		}
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	enemies, err := p.getEnemiesAwards(lvl)
	if err != nil {
		return nil, fmt.Errorf("get awards: %w", err)
	}

	awadrs, punishments, err := parse.ParseCombatLog(
		p.gameLogScanner, p.yourNickname, lvl.LevelEnd,
		func(s string) (int, bool) {
			cost, ok := enemies[s]
			return cost, ok
		})
	if err != nil {
		return nil, fmt.Errorf("parse combat log: %w", err)
	}

	scores := append(awadrs, punishments...)

	return scores, nil
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
