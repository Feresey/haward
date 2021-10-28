package session

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Feresey/haward/parse"
	"github.com/Feresey/haward/rules"
	"go.uber.org/zap"
)

// Parser это сущность которая обрабатывает логи одной сессии
type Parser struct {
	yourNickname string
	rules        *rules.Rules

	levelIter      *parse.GameLogIter
	gameLogScanner *bufio.Scanner

	lastLevel bool
}

func NewParser(
	yourNickname string,
	combat, game io.Reader,
	rules *rules.Rules,
) *Parser {
	return &Parser{
		yourNickname:   yourNickname,
		rules:          rules,
		lastLevel:      false,
		levelIter:      parse.NewGameLogIter(yourNickname, game),
		gameLogScanner: bufio.NewScanner(combat),
	}
}

func (p *Parser) Parse(ctx context.Context, log *zap.Logger, levelReports chan<- *LevelReport) error {
	for !p.lastLevel {
		levelReport, err := p.parseLogLevel(log)
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
func (p *Parser) parseLogLevel(logger *zap.Logger) (levelReport *LevelReport, err error) {
	lvl, err := p.levelIter.ScanNextLevel()
	if err != nil {
		if errors.Is(err, io.EOF) {
			p.lastLevel = true
		}
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	var report LevelReport

	enemies := lvl.GetEnemies()

	logger.Debug("", zap.Reflect("enemies", enemies))

	enemiesAwards, err := p.getEnemiesAwards(enemies)
	if err != nil {
		return nil, fmt.Errorf("get awards: %w", err)
	}
	logger.Debug("", zap.Reflect("enemies_awards", enemiesAwards))

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
		clan, err := p.rules.GetPlayerClan(nickname)
		if err != nil {
			return nil, err
		}
		if clan != nil {
			res[nickname] = Player{
				Player: enemy,
				Clan:   clan.Name,
			}
		}
	}

	return res, nil
}
