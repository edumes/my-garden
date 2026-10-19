package auth

import (
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Experience granted for gameplay actions. A harvest grants the plant's
// growth time in minutes, so slower crops reward more.
const (
	PlantingExperience = 2
	// TradeExperience goes to the seller of a marketplace listing and to the
	// player who completes a delivery request.
	TradeExperience = 10
)

// ExperienceForLevel returns the total experience needed to reach level.
// Going from level n to n+1 costs 100×n, so level 2 needs 100, level 3
// needs 300, and level 10 needs 4500.
func ExperienceForLevel(level int) int {
	return 50 * level * (level - 1)
}

// LevelFor returns the level a player with this much experience has reached.
func LevelFor(experience int) int {
	level := 1
	for ExperienceForLevel(level+1) <= experience {
		level++
	}
	return level
}

// GrantExperience adds experience to the user within tx and raises their
// level to match, returning the resulting level and whether it rose.
func GrantExperience(tx *gorm.DB, userID uuid.UUID, amount int) (int, bool, error) {
	// The update locks the user's row until tx ends, so concurrent grants
	// compute levels one after another.
	var before struct{ Experience, Level int }
	if err := tx.Raw(`UPDATE users SET experience = experience + ? WHERE id = ? RETURNING experience, level`,
		amount, userID).Scan(&before).Error; err != nil {
		return 0, false, err
	}
	level := LevelFor(before.Experience)
	if level <= before.Level {
		return before.Level, false, nil
	}
	err := tx.Model(&User{}).Where("id = ?", userID).Update("level", level).Error
	return level, true, err
}

// MarshalJSON adds the experience bounds of the user's current level, so
// clients show progress without repeating the level curve.
func (u User) MarshalJSON() ([]byte, error) {
	type plain User
	p := plain(u)
	p.LevelExperience = ExperienceForLevel(u.Level)
	p.NextLevelExperience = ExperienceForLevel(u.Level + 1)
	return json.Marshal(p)
}
