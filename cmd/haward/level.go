package main

import (
	"strconv"
	"time"

	"github.com/Feresey/haward/session"
)

type SessionReport struct {
	StartedAt time.Time
	Levels    []*session.LevelReport
}

type SessionIter struct {
	s        *SessionReport
	levelIdx int
	scoreIdx int
}

func NewReportIter(s *SessionReport) *SessionIter {
	return &SessionIter{
		s: s,
	}
}

func (r *SessionIter) Next() bool {
	if len(r.s.Levels) == 0 {
		return false
	}
	r.scoreIdx++
	if r.scoreIdx < len(r.s.Levels[r.levelIdx].Score) {
		return true
	}

	r.scoreIdx = 0
	r.levelIdx++

	return r.levelIdx < len(r.s.Levels)
}

func (r *SessionIter) Header() []string {
	return []string{
		"session_start",
		"killed_at",
		"line_in_log",
		"killed",
		"clan",
		"score",
	}
}

func (r *SessionIter) Line() []string {
	level := r.s.Levels[r.levelIdx]
	line := level.Score[r.scoreIdx]

	return []string{
		r.s.StartedAt.Format(sessionTimeFormat),
		line.Time,
		strconv.Itoa(line.LineNum),
		line.Killed,
		level.Enemies[line.Killed].Clan,
		strconv.Itoa(line.Award),
	}
}
