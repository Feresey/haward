package session

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/Feresey/haward/parse"
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

	rules := Rules{
		awards: map[string]int{
			"lafan4ik":       1,
			"Viktorius31rus": 20,
		},
		punishments: map[string]int{
			"fyringsved": -100,
		},
		clanTags: map[string]int{
			"HPrim": 5,
		},
	}

	p := NewParser("ZiroTwo", combat, game, rules)

	res := make(chan []parse.DeathRecord)
	go func() {
		defer close(res)
		err := p.Parse(res)
		if !errors.Is(err, io.EOF) {
			r.NoError(err)
		}
	}()

	for battle := range res {
		for _, one := range battle {
			fmt.Printf("%s: %d\n", one.Killed, one.Award)
		}
	}
}
