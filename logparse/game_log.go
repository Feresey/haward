package logparse

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type GameLogIter struct {
	rd           *bufio.Reader
	yourNickname string

	levelStarting bool
}

func NewGameLogIter(yourNickname string, r io.Reader) *GameLogIter {
	return &GameLogIter{
		rd:           bufio.NewReader(r),
		yourNickname: yourNickname,
	}
}

type Player struct {
	Name    string
	ID      uint64
	ClanTag string
	InGroup bool
}

type GameLogLevel struct {
	MapName  string
	YourTeam int
	// Players is map[team_id]Player
	Players map[int][]Player
}

func (g *GameLogLevel) GetEnemies() map[string]Player {
	res := make(map[string]Player)
	for team, players := range g.Players {
		if team == g.YourTeam {
			continue
		}

		for _, p := range players {
			res[p.Name] = p
		}
	}
	return res
}

// 12:51:09.342         | ====== starting level: 'levels/area1/s1338_pandora_anomaly' KingOfTheHill client =====
const (
	startingLevelContains = `====== starting level:`
)

func (it *GameLogIter) ScanNextLevel() (*GameLogLevel, error) {
	var lvl GameLogLevel

	for {
		line, err := it.rd.ReadString('\n')
		if err != nil {
			return nil, err
		}

		startingMessage := strings.Contains(line, startingLevelContains)

		if !it.levelStarting {
			// поиск старта уровня, если ещё не было найдено начало
			if startingMessage {
				continue
			}
			it.levelStarting = true
		} else if startingMessage {
			// если начало было найдено, то старт нового уровня означает окончанее предыдущего
			return &lvl, nil
		}

		if err := it.processLogLine(&lvl, line); err != nil {
			return nil, err
		}
	}
}

func (it *GameLogIter) processLogLine(lvl *GameLogLevel, line string) error {
	// 17:27:50.022         | client: ADD_PLAYER 9 (BNV [CSA], 1308282) status 4 team 2 group 4778580
	const (
		addPLayerPrefix    = `ADD_PLAYER`
		playerStatusOnline = 4
		playerStatusKey    = "status"
		playerTeamKey      = "team"
		playerGroupKey     = "group"
	)

	split := strings.Split(line, addPLayerPrefix)

	if len(split) != 2 { // ADD_PLAYER line
		return nil
	}

	addPlayer := split[1]

	playerStart := strings.Index(addPlayer, "(")
	playerEnd := strings.Index(addPlayer, ")")

	if playerStart == -1 || playerEnd == -1 {
		return fmt.Errorf("start: %d, end: %d", playerStart, playerEnd)
	}

	player, err := parsePlayer(addPlayer[playerStart:playerEnd])
	if err != nil {
		return err
	}

	status, team, group, err := it.parseAddPlayerFields(strings.Fields(addPlayer[playerEnd+1:]))
	if err != nil {
		return err
	}

	if player.Name == it.yourNickname {
		lvl.YourTeam = team
	}
	if group != 0 {
		player.InGroup = true
	}
	if status != playerStatusOnline {
		return nil
	}

	lvl.Players[team] = append(lvl.Players[team], *player)
	return nil
}

// (BNV [CSA], 1308282)
func parsePlayer(player string) (*Player, error) {
	fields := strings.Fields(strings.Trim(player, "()"))

	// BNV [CSA], 1308282
	const (
		fieldName = iota
		fieldClanTag
		fieldID
		fieldsLength
	)

	if len(fields) != fieldsLength {
		return nil, fmt.Errorf("could notn parse player: %q", player)
	}

	var p Player
	p.Name = fields[fieldName]
	p.ClanTag = strings.TrimRight(strings.TrimLeft(fields[fieldClanTag], "["), "],")
	id, err := strconv.ParseUint(fields[fieldID], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse player id: %q: %w", fields[fieldID], err)
	}
	p.ID = id
	return &p, nil
}

func (it *GameLogIter) parseAddPlayerFields(fields []string) (status, team, group int, err error) {
	// status 4 team 2 group 4778580
	const (
		addPLayerPrefix    = `ADD_PLAYER`
		playerStatusOnline = 4
		playerStatusKey    = "status"
		playerTeamKey      = "team"
		playerGroupKey     = "group"
	)

	for i := 0; i < len(fields); i += 2 {
		switch fields[i] {
		case playerStatusKey:
			status, err = strconv.Atoi(fields[i+1])
		case playerTeamKey:
			team, err = strconv.Atoi(fields[i+1])
		case playerGroupKey:
			group, err = strconv.Atoi(fields[i+1])
		}
		if err != nil {
			break
		}
	}
	return
}
