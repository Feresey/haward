package main

import (
	"bufio"
	"errors"
	"io"
	"regexp"
)

const (
	fieldTime = iota
	fieldbattleType
	fieldKilledName
	fieldKillerName
	fieldKillWith
	// allFields
)

var killedRe = regexp.MustCompile(`^(?P<time>.*)\s+(?P<battle_type>\w+)\s+\|\s+Killed\s+(?P<killed_name>[[:word:]]+)\s+.*killer\s+(?P<killer_name>[[:word:]]+)\s*\|.*\s+(?P<kill_with>.*)$`)

type DeathRecord struct {
	LineNum  int
	Original string

	Time       string
	BattleType string
	Killed     string
	Killer     string
	KillWith   string

	Award int
}

type Parser struct {
	yourNickname    string
	awardsList      map[string]int
	punishmentsList map[string]int

	playerClanHunt map[string]bool
}

func NewParser(nickname string) *Parser {
	return &Parser{yourNickname: nickname}
}

func (p *Parser) Parse(
	combatLog, gameLog io.Reader,
	awards, punishments io.Writer,
) error {
	const megabyte = 1 << 20

	combatReader := bufio.NewReaderSize(combatLog, megabyte)
	gameReader := bufio.NewReaderSize(gameLog, megabyte)

	

}

func (p *Parser) ParseLog(r io.Reader) (awards, punishments []*DeathRecord, err error) {
	if err := p.validate(); err != nil {
		return nil, nil, err
	}

	var lineNum int

	scanner := bufio.NewScanner(r)
	for ; scanner.Scan(); lineNum++ {
		lineNum++
		line := scanner.Text()

		if !killedRe.MatchString(line) {
			continue
		}

		fields := killedRe.FindStringSubmatch(line)

		record := &DeathRecord{
			LineNum:    lineNum,
			Original:   line,
			Time:       fields[fieldTime],
			BattleType: fields[fieldbattleType],
			Killed:     fields[fieldKilledName],
			Killer:     fields[fieldKillerName],
			KillWith:   fields[fieldKillWith],
		}

		if record.Killer != p.yourNickname {
			continue
		}

		award, ok := p.awardsList[record.Killed]
		if ok {
			record.Award = award
			awards = append(awards, record)
		}

		// TODO add a check if the player was in the group
		punishment, ok := p.punishmentsList[record.Killed]
		if ok {
			record.Award = punishment
			punishments = append(punishments, record)
		}
	}

	return awards, punishments, scanner.Err()
}

func (p *Parser) validate() error {
	if len(p.awardsList) == 0 {
		return errors.New("awards list is empty")
	}
	if len(p.punishmentsList) == 0 {
		return errors.New("punishments list is empty")
	}

	return nil
}

func (p *Parser) ParseAwardsList(r io.Reader) error

func (p *Parser) ParsePunishmentsList(r io.Reader) error

func (p *Parser) parseNcknamesList(r io.Reader) (map[string]int, error) {
	scanner := bufio.NewScanner(r)

	var score int

}
