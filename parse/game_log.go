package parse

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
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

	LevelEnd time.Time
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
		lineBytes, _, err := it.rd.ReadLine()
		line := string(lineBytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return &lvl, err
			}
			return nil, err
		}

		startingMessage := strings.Contains(line, startingLevelContains)

		// если нет сообщения о старте уровня
		if !startingMessage {
			if err := it.processLogLine(&lvl, line); err != nil {
				return nil, err
			}
			continue
		}

		// если логи до этой строчки принадлежали другому уровню
		if it.levelStarting {
			// то уровень завершился и сейчас старт нового
			finishedAt, err := time.Parse(timeFormat, line[:strings.Index(line, " ")])
			if err != nil {
				return nil, fmt.Errorf("parse stop time: %q: %w", line, err)
			}
			lvl.LevelEnd = finishedAt
			return &lvl, nil
		} else { // если логи выше не принадлежали уровню
			// то теперь началось описание уровня
			it.levelStarting = true
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
	playerEnd := strings.LastIndex(addPlayer, ")")

	if playerStart == -1 || playerEnd == -1 {
		return fmt.Errorf("start: %d, end: %d", playerStart, playerEnd)
	}

	player, err := parsePlayer(addPlayer[playerStart : playerEnd+1])
	if err != nil {
		return err
	}

	// бот
	if player.ID == 0 {
		return nil
	}

	status, team, group, err := it.parsePlayerFields(strings.Fields(addPlayer[playerEnd+1:]))
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

	if lvl.Players == nil {
		lvl.Players = make(map[int][]Player)
	}
	lvl.Players[team] = append(lvl.Players[team], *player)
	return nil
}

// (BNV [CSA], 1308282)
func parsePlayer(player string) (*Player, error) {
	fields := strings.Fields(strings.TrimSuffix(strings.TrimPrefix(player, "("), ")"))

	// BNV [CSA], 1308282
	const (
		fieldName = iota
		fieldClanTag
		fieldID
		fieldsLength
	)

	var nameRaw, clanRaw, idRaw string

	if len(fields) != fieldsLength {
		// чел без клана
		if len(fields) != fieldsLength-1 {
			return nil, fmt.Errorf("could not parse player: %q", player)
		}

		nameRaw = strings.TrimRight(fields[fieldName], ",")
		idRaw = fields[fieldClanTag]
	} else {
		nameRaw = fields[fieldName]
		clanRaw = fields[fieldClanTag]
		idRaw = fields[fieldID]
	}

	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse player id: %q: %w", idRaw, err)
	}

	return &Player{
		Name:    nameRaw,
		ID:      id,
		ClanTag: strings.TrimRight(strings.TrimLeft(clanRaw, "["), "],"),
	}, nil
}

func (it *GameLogIter) parsePlayerFields(fields []string) (status, team, group int, err error) {
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
