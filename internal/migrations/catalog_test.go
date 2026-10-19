package migrations

import (
	"math"
	"testing"
)

// The catalog follows the rule documented on plantTypes, so every harvest
// pays back its seed with a profit that grows with the unlock level.
func TestPlantTypesFollowTheEconomy(t *testing.T) {
	levels := map[int]bool{}
	for _, pt := range plantTypes {
		levels[pt.UnlockLevel] = true
		revenue := pt.HarvestValue * pt.Yield
		if revenue <= pt.SeedPrice {
			t.Errorf("%s: a harvest pays %d, not more than its %d-coin seed", pt.Name, revenue, pt.SeedPrice)
		}

		rate := 0.20 + 0.05*float64(pt.UnlockLevel-1)
		want := float64(pt.SeedPrice) + float64(pt.GrowthTime)*rate
		// Rounding HarvestValue moves revenue by up to half a value per item.
		if math.Abs(float64(revenue)-want) > float64(pt.Yield)/2+0.5 {
			t.Errorf("%s: a harvest pays %d, want about %.1f", pt.Name, revenue, want)
		}
	}
	for level := 1; level <= 6; level++ {
		if !levels[level] {
			t.Errorf("no plant type unlocks at level %d", level)
		}
	}
}
