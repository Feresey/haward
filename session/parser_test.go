package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	r := require.New(t)

	combat, err := os.Open("testdata/combat.log")
	r.NoError(err)
	defer combat.Close()

	game, err := os.Open("testdata/game.log")
	r.NoError(err)
	defer game.Close()

	rules := &Rules{
		awards: map[string]int{
			"lafan4ik": 1,
		},
		punishments: map[string]int{
			"fyringsved": -100,
		},
		clanTags: map[string]int{
			"HPrim": 5,
		},
		resolver: NewPlayerResolver(),
	}

	p := NewParser("ZiroTwo", combat, game, rules)
	ctx := context.TODO()

	res := make(chan *LevelReport)
	go func() {
		defer close(res)
		err := p.Parse(ctx, res)
		if !errors.Is(err, io.EOF) {
			r.NoError(err)
		}
	}()

	for battle := range res {
		for _, one := range battle.Score {
			fmt.Printf("%s: %d\n", one.Killed, one.Award)
		}
	}
}
