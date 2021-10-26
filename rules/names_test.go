package rules

import (
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestParseCorpotatoins(t *testing.T) {
	tests := []struct {
		data string
		want [2]string
	}{
		{
			data: "Nekopara",
			want: [2]string{
				"Nekopara", "",
			},
		},
		{
			data: "Feeling of Greatness [4CB]",
			want: [2]string{
				"Feeling of Greatness", "4CB",
			},
		},
		{
			data: "Fright Night [FINS]",
			want: [2]string{
				"Fright Night", "FINS",
			},
		},
		{
			data: "The Dark Invaders [xIDx]",
			want: [2]string{
				"The Dark Invaders", "xIDx",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.data, func(t *testing.T) {
			name, tag := parseCorporation(tt.data)

			require.Equal(t, tt.want, [2]string{name, tag})
		})
	}
}

func TestParseNames(t *testing.T) {
	r := &Rules{
		awards:      make(map[string]int),
		punishments: make(map[string]int),
		clanTags:    make(map[string]int),
		clanNames:   make(map[string]int),
		PlayerClanResolver:    NewPlayerResolver(),
	}

	file, err := os.Open("testdata/names")
	require.NoError(t, err)
	defer file.Close()

	err = r.parseRules(file)
	require.NoError(t, err)

	spew.Dump(r)
}
