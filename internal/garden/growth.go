package garden

import (
	"time"

	"gorm.io/gorm"
)

// StageFor returns the stage a plant shows at a growth progress from 0 to 100.
func StageFor(progress float64) PlantStage {
	switch {
	case progress >= 100:
		return PlantStageHarvestable
	case progress >= 70:
		return PlantStageMature
	case progress >= 40:
		return PlantStageGrowing
	case progress >= 20:
		return PlantStageSprout
	default:
		return PlantStageSeed
	}
}

// Grow returns the progress of a plant that ripens in growthMinutes after it
// grows for elapsed time. The weather multiplier and the game speed scale the
// rate, so at multiplier 1 and speed 1 the plant ripens in exactly
// growthMinutes.
func Grow(progress float64, growthMinutes int, elapsed time.Duration, multiplier, speed float64) float64 {
	progress += 100 * elapsed.Minutes() * multiplier * speed / float64(growthMinutes)
	return min(progress, 100)
}

// GrowPlants advances every unripe plant from when it last grew until now.
// Progress and grown_at are written together from one read, so each interval
// grows a plant once however ticks overlap, run on several instances, or
// repeat after a restart. Each write is also conditional on the grown_at it
// read, so a slow tick never rolls back a newer one. Only growth columns are
// written: a plant moved meanwhile stays put, and a harvested or removed one
// stays gone.
func GrowPlants(db *gorm.DB, now time.Time, multiplier, speed float64) error {
	var plants []Plant
	if err := db.Preload("PlantType").
		Where("stage <> ? AND grown_at < ?", PlantStageHarvestable, now).
		Find(&plants).Error; err != nil {
		return err
	}

	for _, plant := range plants {
		progress := Grow(plant.GrowthProgress, plant.PlantType.GrowthTime, now.Sub(plant.GrownAt), multiplier, speed)
		if err := db.Model(&Plant{}).
			Where("id = ? AND grown_at = ?", plant.ID, plant.GrownAt).
			Updates(map[string]any{"growth_progress": progress, "stage": StageFor(progress), "grown_at": now}).Error; err != nil {
			return err
		}
	}
	return nil
}
