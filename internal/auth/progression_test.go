package auth

import (
	"encoding/json"
	"testing"
)

func TestLevelCurve(t *testing.T) {
	cases := map[int]int{0: 1, 99: 1, 100: 2, 299: 2, 300: 3, 1000: 5, 4499: 9, 4500: 10}
	for experience, want := range cases {
		if got := LevelFor(experience); got != want {
			t.Errorf("LevelFor(%d) = %d, want %d", experience, got, want)
		}
	}
	for level := 1; level < 20; level++ {
		if cost := ExperienceForLevel(level+1) - ExperienceForLevel(level); cost != 100*level {
			t.Errorf("level %d to %d costs %d, want %d", level, level+1, cost, 100*level)
		}
	}
}

func TestUserJSONIncludesLevelBounds(t *testing.T) {
	raw, err := json.Marshal(User{Level: 3, Experience: 350})
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		LevelExperience     int `json:"level_experience"`
		NextLevelExperience int `json:"next_level_experience"`
		GardenLimit         int `json:"garden_limit"`
		NextGardenLevel     int `json:"next_garden_level"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.LevelExperience != 300 || got.NextLevelExperience != 600 {
		t.Errorf("got bounds %+v, want 300 and 600", got)
	}
	if got.GardenLimit != 2 || got.NextGardenLevel != 5 {
		t.Errorf("got garden unlocks %+v, want limit 2 and next level 5", got)
	}
}

func TestGardenLimitsMatchProfile(t *testing.T) {
	for level, want := range map[int]int{1: 1, 2: 1, 3: 2, 4: 2, 5: 3, 20: 10} {
		if got := MaxGardens(level); got != want {
			t.Errorf("MaxGardens(%d) = %d, want %d", level, got, want)
		}
	}
}
