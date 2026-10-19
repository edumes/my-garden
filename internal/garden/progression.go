package garden

import (
	"fmt"

	"github.com/google/uuid"
)

// MaxGardens is how many gardens a player of this level may own: one at
// levels 1 and 2, then one more every two levels. Shared gardens do not count.
func MaxGardens(level int) int {
	return (level + 1) / 2
}

// gardenLimitError reports that the owner needs a higher level for another
// garden.
type gardenLimitError struct{ owned, nextLevel int }

func (e gardenLimitError) Error() string {
	gardens := "gardens"
	if e.owned == 1 {
		gardens = "garden"
	}
	return fmt.Sprintf("you own %d %s, the most your level allows; reach level %d to create another", e.owned, gardens, e.nextLevel)
}

// Reward tells the client what an action paid and to whom.
type Reward struct {
	UserID     uuid.UUID `json:"user_id"`
	Coins      int       `json:"coins"`
	Experience int       `json:"experience"`
	Level      int       `json:"level"`
	LevelUp    bool      `json:"level_up"`
}
