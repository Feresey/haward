package parse

import (
	"bufio"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegexp(t *testing.T) {
	lines := []string{
		"21:08:54.870  CMBT   | Killed NikSvir	 Ship_Race2_S_T3_Premium|0000002708;	 killer ZiroTwo|0000002012 Weapon_Railgun_Sniper_T4_Rel",
		"21:11:39.922  CMBT   | Killed NikSvir	 Ship_Race2_S_T3_Premium|0000125243;	 killer ZiroTwo|0000128312 SpaceMissile_Torpedo_T3_Mk3",
		"20:40:23.253  CMBT   | Killed HoWHoW	 Ship_Race1_M_T5_Faction2|0000003396;	 killer ZiroTwo|0000000374 Module_GuidedMissile_T4_Base",
	}

	for num, line := range lines {
		res := killedRe.MatchString(line)
		require.True(t, res, "line %d: %q", num, line)
	}
}

func TestParseCombatLog(t *testing.T) {
	t.Run("part", func(t *testing.T) {
		r := require.New(t)
		file, err := os.Open("testdata/combat.log")
		r.NoError(err)
		defer file.Close()
		// 21:53:19.835
		awards, punishments, err := ParseCombatLog(
			bufio.NewScanner(file),
			"ZiroTwo",
			time.Date(2021, time.October, 19, 22, 0, 0, 0, time.Local),
			func(s string) (int, bool) {
				if s == "NikSvir" {
					return 10, true
				}
				if s == "Inspiration" {
					return -40, true
				}
				return 0, false
			})
		r.NoError(err)

		r.Equal(
			[]DeathRecord{
				{
					LineNum:  8419,
					Original: "21:42:29.979  CMBT   | Killed NikSvir\t Ship_Race2_S_T3_Premium|0000002478;\t killer ZiroTwo|0000058934 Weapon_Railgun_Sniper_T4_Rel ",
					Time:     "21:42:29.979",
					Killed:   "NikSvir",
					Killer:   "ZiroTwo",
					KillWith: "Weapon_Railgun_Sniper_T4_Rel",
					Award:    10,
				},
				{
					LineNum:  38112,
					Original: "21:46:31.864  CMBT   | Killed NikSvir\t Ship_Race2_S_T3_Premium|0000187861;\t killer ZiroTwo|0000058934 Weapon_Railgun_Sniper_T4_Rel ",
					Time:     "21:46:31.864",
					Killed:   "NikSvir",
					Killer:   "ZiroTwo",
					KillWith: "Weapon_Railgun_Sniper_T4_Rel",
					Award:    10,
				},
				{
					LineNum:  57820,
					Original: "21:53:19.834  CMBT   | Killed NikSvir\t Ship_Race5_L_ENGINEER_Rank9_7|0000002801;\t killer ZiroTwo|0000028238 Weapon_Railgun_Sniper_T4_Rel ",
					Time:     "21:53:19.834",
					Killed:   "NikSvir",
					Killer:   "ZiroTwo",
					KillWith: "Weapon_Railgun_Sniper_T4_Rel",
					Award:    10,
				},
			},
			awards,
			"awards",
		)

		r.Equal(
			[]DeathRecord{
				{
					LineNum:  19159,
					Original: "21:44:00.880  CMBT   | Killed Inspiration\t Ship_Race1_M_T3_Faction3|0000149939;\t killer ZiroTwo|0000058934 Weapon_Railgun_Sniper_T4_Rel ",
					Time:     "21:44:00.880",
					Killed:   "Inspiration",
					Killer:   "ZiroTwo",
					KillWith: "Weapon_Railgun_Sniper_T4_Rel",
					Award:    -40,
				},
			},
			punishments,
			"punishments",
		)
	})

	t.Run("parse full log", func(t *testing.T) {
		r := require.New(t)
		file, err := os.Open("testdata/combat_full.log")
		r.NoError(err)
		defer file.Close()

		awards, punishments, err := ParseCombatLog(
			bufio.NewScanner(file),
			"ZiroTwo",
			time.Date(2021, time.October, 19, 23, 0, 0, 0, time.Local),
			func(s string) (int, bool) {
				if s == "PromptoArgentum" {
					return 10, true
				}
				if s == "Theokles" {
					return -40, true
				}
				return 0, false
			})
		r.NoError(err)

		r.Len(awards, 1)
		r.Len(punishments, 3)
	})
}
