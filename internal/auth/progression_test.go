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
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.LevelExperience != 300 || got.NextLevelExperience != 600 {
		t.Errorf("got bounds %+v, want 300 and 600", got)
	}
}
