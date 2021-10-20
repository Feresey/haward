package main

import (
	"strconv"
	"time"

	"github.com/Feresey/haward/parse"
)

type LevelReport struct {
	Date  time.Time
	Score []parse.DeathRecord
}

type LevelReportIter struct {
	l   *LevelReport
	idx int
}

func NewLevelReportIter(l *LevelReport) *LevelReportIter {
	return &LevelReportIter{
		l:   l,
		idx: -1,
	}
}

func (r *LevelReportIter) Next() { r.idx++ }

func (r *LevelReportIter) Line() []string {
	line := r.l.Score[r.idx]

	return []string{
		r.l.Date.Format(time.Stamp),
		line.Time,
		strconv.Itoa(line.LineNum),
		line.Killed,
		strconv.Itoa(line.Award),
	}
}
