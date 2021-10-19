package parse

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePlayer(t *testing.T) {
	tests := []struct {
		data string
		want *Player
	}{
		{
			data: "(ZiroTwo [xIDx], 2516405)",
			want: &Player{
				Name:    "ZiroTwo",
				ClanTag: "xIDx",
				ID:      2516405,
			},
		},
		{
			data: "(Gob [FlyAR], 3767922)",
			want: &Player{
				Name:    "Gob",
				ClanTag: "FlyAR",
				ID:      3767922,
			},
		},
		{
			data: "(Dimon856 [xIDx], 284392)",
			want: &Player{
				Name:    "Dimon856",
				ClanTag: "xIDx",
				ID:      284392,
			},
		},
		{
			data: "(Walle00one [FlyAR], 3748346)",
			want: &Player{
				Name:    "Walle00one",
				ClanTag: "FlyAR",
				ID:      3748346,
			},
		},
		{
			data: "(ZiroTwo [xIDx], 2516405)",
			want: &Player{
				Name:    "ZiroTwo",
				ClanTag: "xIDx",
				ID:      2516405,
			},
		},
		{
			data: "(HemitM [], 2020202)",
			want: &Player{
				Name:    "HemitM",
				ClanTag: "",
				ID:      2020202,
			},
		},
		{
			data: "(AlKorn, 0)",
			want: &Player{
				Name:    "AlKorn",
				ClanTag: "",
				ID:      0,
			},
		},
	}

	for _, test := range tests {
		tt := test
		t.Run(test.data, func(t *testing.T) {
			player, err := parsePlayer(tt.data)
			require.NoError(t, err)
			require.Equal(t, tt.want, player)
		})
	}
}

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		data string
		want *GameLogLevel
	}{
		{
			data: "12:51:10.311         | client: ADD_PLAYER 0 (ZiroTwo [xIDx], 2516405) status 2 team 1 group 4778580",
			want: &GameLogLevel{
				YourTeam: 1,
				Players:  make(map[int][]Player),
			},
		},
		{
			data: "12:51:10.315         | client: ADD_PLAYER 0 (ZiroTwo [xIDx], 2516405) status 4 team 1 group 4778580",
			want: &GameLogLevel{
				YourTeam: 1,
				Players: map[int][]Player{
					1: {{
						Name:    "ZiroTwo",
						ClanTag: "xIDx",
						ID:      2516405,
						InGroup: true,
					}},
				},
			},
		},
		{
			data: "12:51:10.311         | client: ADD_PLAYER 1 (Gob [FlyAR], 3767922) status 4 team 2",
			want: &GameLogLevel{
				Players: map[int][]Player{
					2: {{
						Name:    "Gob",
						ClanTag: "FlyAR",
						ID:      3767922,
						InGroup: false,
					}},
				},
			},
		},
		{
			data: "12:51:10.312         | client: ADD_PLAYER 2 (Dimon856 [xIDx], 284392) status 4 team 1 group 4778580",
			want: &GameLogLevel{
				Players: map[int][]Player{
					1: {{
						Name:    "Dimon856",
						ClanTag: "xIDx",
						ID:      284392,
						InGroup: true,
					}},
				},
			},
		},
		{
			data: "12:51:10.312         | client: ADD_PLAYER 3 (Walle00one [FlyAR], 3748346) status 4 team 2 group 4778574",
			want: &GameLogLevel{
				Players: map[int][]Player{
					2: {{
						Name:    "Walle00one",
						ClanTag: "FlyAR",
						ID:      3748346,
						InGroup: true,
					}},
				},
			},
		},
	}

	for _, test := range tests {
		tt := test

		t.Run("", func(t *testing.T) {
			li := &GameLogIter{
				yourNickname:  "ZiroTwo",
				rd:            nil,
				levelStarting: false,
			}

			var lvl GameLogLevel
			lvl.Players = make(map[int][]Player)
			err := li.processLogLine(&lvl, test.data)
			require.NoError(t, err)
			tt.want.LevelEnd = lvl.LevelEnd
			require.Equal(t, tt.want, &lvl)
		})
	}
}

func TestParseGameLog(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		r := require.New(t)
		file, err := os.Open("testdata/game_empty.log")
		r.NoError(err)
		defer file.Close()

		gameLog := NewGameLogIter("ZiroTwo", file)

		level, err := gameLog.ScanNextLevel()
		r.EqualError(err, io.EOF.Error())

		r.True(level.LevelEnd.IsZero())
		r.Equal(
			&GameLogLevel{},
			level,
		)
	})

	t.Run("one", func(t *testing.T) {
		r := require.New(t)
		file, err := os.Open("testdata/game_one.log")
		r.NoError(err)
		defer file.Close()

		gameLog := NewGameLogIter("ZiroTwo", file)

		level, err := gameLog.ScanNextLevel()
		r.NoError(err)
		r.False(level.LevelEnd.IsZero())
		r.Equal(
			&GameLogLevel{LevelEnd: level.LevelEnd},
			level,
		)

		level, err = gameLog.ScanNextLevel()
		r.NoError(err)
		r.False(level.LevelEnd.IsZero())
		r.Equal(
			&GameLogLevel{
				LevelEnd: level.LevelEnd,
				YourTeam: 1,
				Players: map[int][]Player{
					1: {
						{
							Name:    "ZiroTwo",
							ID:      2516405,
							ClanTag: "xIDx",
							InGroup: true,
						},
						{
							Name:    "Dimon856",
							ID:      284392,
							ClanTag: "xIDx",
							InGroup: true,
						},
					},
					2: {
						{
							Name:    "Gob",
							ID:      3767922,
							ClanTag: "",
							InGroup: true,
						},
						{
							Name:    "Walle00one",
							ID:      3748346,
							ClanTag: "",
							InGroup: true,
						},
					},
				},
			},
			level,
		)

		level, err = gameLog.ScanNextLevel()
		r.EqualError(err, io.EOF.Error())
		r.True(level.LevelEnd.IsZero())
		r.Equal(
			&GameLogLevel{LevelEnd: level.LevelEnd},
			level,
		)
	})

	t.Run("all", func(t *testing.T) {
		r := require.New(t)
		file, err := os.Open("testdata/game_all.log")
		r.NoError(err)
		defer file.Close()

		gameLog := NewGameLogIter("ZiroTwo", file)

		counts := []int{
			0, 0, 4, 0, 4,
			2, 4, 2, 4, 0,
			4, 0, 4, 0, 4,
			0, 4, 1, 4, 0,
			4, 0, 4, 0, 4,
			0, 3, 0, 4, 0,
			8, 3, 8, 0, 8,
			0, 7, 0, 6, 0,
			6, 0, 6, 0, 8,
			0, 4, 0, 8, 3,
			8,
		}

		for count := 0; ; count++ {
			level, err := gameLog.ScanNextLevel()
			if errors.Is(err, io.EOF) {
				break
			}
			r.NoError(err)

			r.False(level.LevelEnd.IsZero())
			r.Len(level.GetEnemies(), counts[count])
		}
	})
}
