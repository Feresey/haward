package main

import "github.com/Feresey/haward/parse"

type CombatLogParser struct {
	yourNickname string
	// награды за конкретных пилотов
	awardsList map[string]int
	// повелители бури
	punishmentsList map[string]int

	// информация о людях в бою
	level *parse.GameLogLevel

	//
	// playerClanHunt map[string]bool
}

func NewCombatLogParser(
	nickname string,
	awards, punishments map[string]int,
) *CombatLogParser {
	return &CombatLogParser{yourNickname: nickname}
}
