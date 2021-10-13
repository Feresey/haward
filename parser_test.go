package main

import (
	"testing"

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
