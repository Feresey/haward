package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Feresey/haward/rules"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParse(t *testing.T) {
	r := require.New(t)

	combat, err := os.Open("testdata/combat.log")
	r.NoError(err)
	defer combat.Close()

	game, err := os.Open("testdata/game.log")
	r.NoError(err)
	defer game.Close()

	const rulesTxt = `
=== PLAYERS ===
+1
lafan4ik
-100
fyringsved
=== CORPORATIONS ===
+5 HPrim
`

	rule, err := rules.NewRules(strings.NewReader(rulesTxt))
	r.NoError(err)

	p := NewParser("ZiroTwo", combat, game, rule)
	ctx := context.TODO()

	res := make(chan *LevelReport)
	go func() {
		defer close(res)
		err := p.Parse(ctx, zap.NewNop(), res)
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
