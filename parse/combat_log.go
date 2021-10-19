package parse

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

const (
	fieldTime = iota + 1
	fieldKilledName
	fieldKillerName
	fieldKillWith
	allFields
)

var killedRe = regexp.MustCompile(`^(?P<time>\S+)\s+CMBT\s+\|\s+Killed\s+(?P<killed_name>\S+)\s+\S+\|\d+\;\s+killer\s+(?P<killer_name>\S+)\|\d+\s+(?P<kill_with>\S*)\s*$`)

type DeathRecord struct {
	LineNum  int
	Original string

	Time     string
	Killed   string
	Killer   string
	KillWith string

	Award int
}

// ParseCombatLog достаёт из лога информацию об убийствах до определённого времени (коцна боя по идее)
func ParseCombatLog(
	r io.Reader,
	yourNickname string,
	until time.Time,
	checkAward func(string) (int, bool),
) (awards, punishments []*DeathRecord, err error) {
	scanner := bufio.NewScanner(r)
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := scanner.Text()

		if !strings.Contains(line, "killer "+yourNickname) {
			continue
		}

		fields := killedRe.FindStringSubmatch(line)
		if len(fields) != allFields {
			// ну тут явно не убийство игрока
			continue
		}

		record := &DeathRecord{
			LineNum:  lineNum,
			Original: line,
			Time:     fields[fieldTime],
			Killed:   fields[fieldKilledName],
			Killer:   fields[fieldKillerName],
			KillWith: fields[fieldKillWith],
		}

		if record.Killer != yourNickname {
			continue
		}

		// TODO add a check if the player was in the group
		award, ok := checkAward(record.Killed)
		if !ok {
			continue
		}

		record.Award = award
		if award > 0 {
			awards = append(awards, record)
		} else {
			punishments = append(punishments, record)
		}
	}

	err = scanner.Err()
	if errors.Is(err, io.EOF) {
		return nil, nil, err
	}

	return awards, punishments, err
}
