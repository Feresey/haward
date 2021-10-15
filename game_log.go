package main

import (
	"bufio"
	"fmt"
	"strings"
)

type GameLogIter struct {
	rd *bufio.Reader	
	yourNickname string	

	levelStarting bool
}

type Player struct {
	Name    string
	ClanTag string
	InGroup bool
}

type GameLogLevel struct {
	MapName    string
	yourTeam int
	Enemies map[string]Player
}

const (
	startingLevelContains = `====== starting level:`
	addPLayerPrefix       = `ADD_PLAYER`
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

	player, err := it.parsePlayer(addPlayer[playerStart:playerEnd])
	if err != nil {
		return err
	}

	fields := 
}

func (it *GameLogIter) parsePlayer(player string) (Player, error)
