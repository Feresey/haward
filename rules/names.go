package rules

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

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

	*PlayerClanResolver
}

func NewRules(rd io.Reader) (*Rules, error) {
	resolver := NewPlayerResolver()

	r := &Rules{
		awards:             make(map[string]int),
		punishments:        make(map[string]int),
		clanTags:           make(map[string]int),
		clanNames:          make(map[string]int),
		PlayerClanResolver: resolver,
	}

	return r, r.parseRules(rd)
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
		clan, err := r.PlayerClanResolver.GetPlayerClan(player.Name)
		if err != nil {
			return 0, false
		}
		award, ok = r.clanNames[clan.Name]
		if ok {
			return
		}
	}
	return 0, false
}

func (r *Rules) parseRules(rd io.Reader) error {
	scanner := bufio.NewScanner(rd)

	var (
		chapter   string
		needScore = true
		score     int
	)

	const (
		chapterPlayers      = "=== PLAYERS ==="
		chapterCorporations = "=== CORPORATIONS ==="
		scoreDelim          = "==="
	)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// chapters
		switch line {
		case chapterPlayers, chapterCorporations:
			chapter = line
			fallthrough
		case scoreDelim:
			needScore = true
			continue
		}

		// score number
		if needScore {
			num, err := strconv.ParseInt(line, 10, 32)
			if err != nil {
				return fmt.Errorf("parse num: %q: %w", num, err)
			}
			score = int(num)
			needScore = false
			continue
		}

		var err error

		// names
		switch chapter {
		case chapterPlayers:
			err = r.parsePlayer(line, score)
		case chapterCorporations:
			name, tag := parseCorporation(line)
			if tag != "" {
				r.clanTags[tag] = score
			}
			r.clanNames[name] = score
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Rules) parsePlayer(s string, score int) error {
	names := strings.Split(s, ",")
	for idx := range names {
		names[idx] = strings.TrimPrefix(names[idx], " ")
	}

	if len(names) == 0 {
		return fmt.Errorf("nickname not found: %q", s)
	}

	setAward := func(line string) {
		if score > 0 {
			r.awards[line] = score
		} else {
			r.punishments[line] = score
		}
	}

	if len(names) == 1 {
		setAward(names[0])
	}

	currName := names[len(names)-1]
	for i := 0; i < len(names)-1; i++ {
		setAward(names[i])
		err := r.PlayerClanResolver.AddOldNickname(names[i], currName)
		if err != nil {
			return err
		}
	}
	setAward(currName)

	return nil
}

func parseCorporation(s string) (string, string) {
	tagBegin := strings.Index(s, "[")
	tagEnd := strings.LastIndex(s, "]")

	if tagBegin != -1 && tagEnd != -1 {
		return s[:tagBegin-1], s[tagBegin+1 : tagEnd]
	}

	return s, ""
}
