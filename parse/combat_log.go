package parse

import (
	"bufio"
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
	scanner *bufio.Scanner,
	yourNickname string,
	until time.Time,
	checkAward func(string) (int, bool),
) (awards, punishments []DeathRecord, err error) {
	killerLine := "killer " + yourNickname

	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := scanner.Text()

		if checkAfter(line, until) {
			break
		}
		if !strings.Contains(line, killerLine) {
			continue
		}

		fields := killedRe.FindStringSubmatch(line)
		if len(fields) != allFields {
			// ну тут явно не убийство игрока
			continue
		}

		record := DeathRecord{
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

	return awards, punishments, err
}

func checkAfter(line string, until time.Time) bool {
	idx := strings.Index(line, " ")
	if idx == -1 {
		return false
	}
	logTime, err := time.Parse(timeFormat, line[:idx])
	if err == nil {
		if logTime.After(until) {
			return true
		}
	}
	return false
}
