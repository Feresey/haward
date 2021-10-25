package session

import (
	"bufio"
	"context"
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

func NewRules(path string) (*Rules, error) {
	resolver := NewPlayerResolver()

	// TODO
	return &Rules{
		awards: map[string]int{
			"Inspiration":    12,
			"NikSvir":        12,
			"NorthMetalhead": 42,
		},
		punishments: map[string]int{
			"Mzelskii": -100500,
		},
		clanTags: map[string]int{
			"4CB": 10,
		},
		clanNames: map[string]int{
			"Nekopara": 10,
		},
		resolver: resolver,
	}, nil
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

// Parser это сущность которая обрабатывает логи одной сессии
type Parser struct {
	yourNickname string
	rules        *Rules

	levelIter      *parse.GameLogIter
	gameLogScanner *bufio.Scanner

	lastLevel bool
}

func NewParser(
	yourNickname string,
	combat, game io.Reader,
	rules *Rules,
) *Parser {
	return &Parser{
		yourNickname:   yourNickname,
		rules:          rules,
		lastLevel:      false,
		levelIter:      parse.NewGameLogIter(yourNickname, game),
		gameLogScanner: bufio.NewScanner(combat),
	}
}

func (p *Parser) Parse(ctx context.Context, levelReports chan<- *LevelReport) error {
	for !p.lastLevel {
		levelReport, err := p.parseLogLevel()
		if err != nil {
			return err
		}

		select {
		case levelReports <- levelReport:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

type Player struct {
	parse.Player
	// Clan is the name of the clan or its tag
	Clan string
}

type LevelReport struct {
	Enemies map[string]Player
	Score   []parse.DeathRecord
}

// parseLogLevel парсит один уровень (одну игру по идее)
func (p *Parser) parseLogLevel() (levelReport *LevelReport, err error) {
	lvl, err := p.levelIter.ScanNextLevel()
	if err != nil {
		if errors.Is(err, io.EOF) {
			p.lastLevel = true
		}
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	var report LevelReport

	enemies := lvl.GetEnemies()

	enemiesAwards, err := p.getEnemiesAwards(enemies)
	if err != nil {
		return nil, fmt.Errorf("get awards: %w", err)
	}

	awadrs, punishments, err := parse.ParseCombatLog(
		p.gameLogScanner, p.yourNickname, lvl.LevelEnd,
		func(s string) (int, bool) {
			cost, ok := enemiesAwards[s]
			return cost, ok
		})
	if err != nil {
		return nil, fmt.Errorf("parse combat log: %w", err)
	}

	report.Score = append(awadrs, punishments...)

	enemiesExtended, err := p.getEnemiesExtended(enemies)
	if err != nil {
		return nil, fmt.Errorf("get enemies extended: %w", err)
	}

	report.Enemies = enemiesExtended

	return &report, nil
}

func (p *Parser) getEnemiesAwards(enemies map[string]parse.Player) (map[string]int, error) {
	awards := make(map[string]int)
	for _, enemy := range enemies {
		award, ok := p.rules.GetAward(enemy)
		if ok {
			awards[enemy.Name] = award
			continue
		}
	}
	return awards, nil
}

func (p *Parser) getEnemiesExtended(enemies map[string]parse.Player) (map[string]Player, error) {
	res := make(map[string]Player)
	for nickname, enemy := range enemies {
		if enemy.ClanTag != "" {
			res[nickname] = Player{
				Player: enemy,
				Clan:   enemy.ClanTag,
			}
			continue
		}
		clanName, err := p.rules.resolver.GetPlayerClanName(nickname)
		if err != nil {
			return nil, err
		}
		if clanName != "" {
			res[nickname] = Player{
				Player: enemy,
				Clan:   clanName,
			}
		}
	}

	return res, nil
}
