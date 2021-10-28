package main

import (
	"testing"
	"time"

	"github.com/Feresey/haward/parse"
	"github.com/Feresey/haward/session"
	"github.com/stretchr/testify/require"
)

func TestLevelIter(t *testing.T) {
	lvl := &SessionReport{
		StartedAt: time.Date(2021, time.December, 0, 0, 0, 0, 0, time.Local),
		Levels: []*session.LevelReport{
			{
				Enemies: map[string]session.Player{
					"first": {
						Clan: "clan",
						Player: parse.Player{
							Name:    "first",
							ID:      007,
							ClanTag: "tag",
							InGroup: false,
						},
					},
					"second": {
						Clan: "clan",
						Player: parse.Player{
							Name:    "second",
							ID:      006,
							ClanTag: "tag",
							InGroup: false,
						},
					},
				},
				Score: []parse.DeathRecord{
					{
						LineNum:  1,
						Original: "log line",
						Time:     "time line",
						Killed:   "first",
						Killer:   "me",
						KillWith: "bonk",
						Award:    42,
					},
					{
						LineNum:  2,
						Original: "log line",
						Time:     "time line",
						Killed:   "second",
						Killer:   "me",
						KillWith: "bonk",
						Award:    43,
					},
				},
			},
		},
	}

	it := NewReportIter(lvl)

	require.True(t, it.Next())
	require.Equal(t, []string{"2021.11.30 00.00.00", "time line", "1", "first", "clan", "42"}, it.Line())

	require.True(t, it.Next())
	require.Equal(t, []string{"2021.11.30 00.00.00", "time line", "2", "second", "clan", "43"}, it.Line())

	require.False(t, it.Next())

	t.Run("empty session", func(t *testing.T) {
		lvl := &SessionReport{
			StartedAt: time.Time{},
			Levels:    nil,
		}

		it := NewReportIter(lvl)

		require.False(t, it.Next())
	})

	t.Run("empty level", func(t *testing.T) {
		lvl := &SessionReport{
			StartedAt: time.Time{},
			Levels: []*session.LevelReport{
				{
					Enemies: nil,
					Score:   nil,
				},
			},
		}

		it := NewReportIter(lvl)

		require.False(t, it.Next())
	})

	t.Run("empty second level", func(t *testing.T) {
		lvl := &SessionReport{
			StartedAt: time.Time{},
			Levels: []*session.LevelReport{
				{
					Enemies: nil,
					Score:   nil,
				},
				{
					Enemies: nil,
					Score:   nil,
				},
			},
		}

		it := NewReportIter(lvl)

		require.True(t, it.Next())
		require.False(t, it.Next())
	})
}
