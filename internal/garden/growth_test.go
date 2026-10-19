package garden

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStageBoundaries(t *testing.T) {
	cases := map[float64]PlantStage{
		0:     PlantStageSeed,
		19.99: PlantStageSeed,
		20:    PlantStageSprout,
		39.99: PlantStageSprout,
		40:    PlantStageGrowing,
		69.99: PlantStageGrowing,
		70:    PlantStageMature,
		99.99: PlantStageMature,
		100:   PlantStageHarvestable,
	}
	for progress, want := range cases {
		assert.Equal(t, want, StageFor(progress), "progress %v", progress)
	}
}

func TestGrowRipensInGrowthTime(t *testing.T) {
	// A 30-minute plant in normal weather at real-time speed.
	assert.InDelta(t, 50, Grow(0, 30, 15*time.Minute, 1, 1), 1e-9)
	assert.Equal(t, 100.0, Grow(0, 30, 30*time.Minute, 1, 1))
	assert.Equal(t, 100.0, Grow(90, 30, time.Hour, 1, 1), "progress stops at 100")
}

func TestGrowScalesWithWeatherAndSpeed(t *testing.T) {
	assert.InDelta(t, 60, Grow(0, 30, 15*time.Minute, 1.2, 1), 1e-9, "sunny weather grows 20% faster")
	assert.InDelta(t, 25, Grow(0, 30, 15*time.Minute, 0.5, 1), 1e-9, "snow halves growth")
	assert.Equal(t, 100.0, Grow(0, 30, 30*time.Second, 1, 60), "GAME_SPEED=60 turns minutes into seconds")
}

func TestGrowIsAdditive(t *testing.T) {
	// Ticks of any length add up to the same progress, so how often the engine
	// runs does not change how fast plants grow.
	oneStep := Grow(0, 45, 20*time.Minute, 1.1, 1)
	manySteps := 0.0
	for range 20 {
		manySteps = Grow(manySteps, 45, time.Minute, 1.1, 1)
	}
	assert.InDelta(t, oneStep, manySteps, 1e-9)
}
