package main

import (
	"bufio"
	"io"
	"regexp"
)

const (
	fieldTime = iota
	fieldbattleType
	fieldKilledName
	fieldKillerName
	fieldKillWith
	allFields
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
	yourNickname string
	awardsList   map[string]int
}

func (p *Parser) ParseLog(r io.Reader) (awards, punishments []DeathRecord, err error) {
	var lineNum int

	for rd := bufio.NewScanner(r); rd.Scan(); lineNum++ {
		lineNum++
		line := rd.Text()

		if !killedRe.MatchString(line) {
			continue
		}

		fields := killedRe.FindStringSubmatch(line)
		killed := fields[fieldKilledName]

		cost, ok := p.getAward(killed)
		iferr

		kill := &DeathRecord{
			LineNum:    lineNum,
			Original:   line,
			Time:       fields[fieldTime],
			BattleType: fields[fieldbattleType],
			Killed:     killed,
			Killer:     fields[fieldKillerName],
			KillWith:   fields[fieldKillWith],
		}
	}
}

func (p *Parser) getAward(name string) (int, bool)

func (p *Parser) ParseAwardsList(r io.Reader) error
