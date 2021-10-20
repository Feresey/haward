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

	resolver *PlayerClanResolver
}

func (r *Rules) GetAward(player parse.Player) (award int, ok bool) {
	award, ok = r.awards[player.Name]
	if ok {
		return
	}
	award, ok = r.punishments[player.Name]
	if ok {
		// повелителей бури можно сбивать если они в группе
		if player.InGroup {
			return 0, false
		}
		return
	}
	award, ok = r.clanTags[player.ClanTag]
	if ok {
		return
	}

	if player.ClanTag == "" {
		fullClanName, err := r.resolver.GetPlayerClanName(player.Name)
		if err != nil {
			return 0, false
		}
		award, ok = r.clanNames[fullClanName]
		if ok {
			return
		}
	}
	return 0, false
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

// Parser это сущность которая обрабатывает логи одной сессии
type Parser struct {
	yourNickname string
	rules        Rules

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
		award, ok := p.rules.GetAward(enemy)
		if ok {
			awards[enemy.Name] = award
			continue
		}
	}
	return awards, nil
}
