package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m4rk1sov/rbk-api/internal/domain/models"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// url link for API, wger
// 2 options, 1st: var init with .env; 2nd: const init
var wgerURL = os.Getenv("WGER_URL")

//const wgerURL = "https://wger.de/api/v2"

var httpClient = &http.Client{Timeout: 5 * time.Second}
var cache = sync.Map{} // key:value

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

func cacheGet(key string) (interface{}, bool) {
	if v, ok := cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Since(entry.timestamp) < 5*time.Minute {
			return entry.data, true
		}
	}
	return nil, false
}

func cacheSet(key string, value interface{}) {
	cache.Store(key, cacheEntry{data: value, timestamp: time.Now().Local()})
}

func GetMuscleID(muscleName string) (int, error) {
	if v, ok := cacheGet("muscles"); ok {
		return findMuscleID(v.([]models.Muscle), muscleName)
	}

	url := wgerURL + "/muscle/"
	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		if closeErr := Body.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
			return
		}
	}(resp.Body)

	var data models.MuscleResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	cacheSet("muscles", data.Results)
	return findMuscleID(data.Results, muscleName)
}

func findMuscleID(muscles []models.Muscle, muscleName string) (int, error) {
	for _, m := range muscles {
		if strings.EqualFold(m.Name, muscleName) {
			return m.ID, nil
		}
	}

	return 0, fmt.Errorf("muscle not found")
}

func GetExercisesByMuscle(muscleID int) ([]models.Exercise, error) {
	cacheKey := fmt.Sprintf("exercises_%d", muscleID)
	if v, ok := cacheGet(cacheKey); ok {
		return v.([]models.Exercise), nil
	}

	url := fmt.Sprintf("%s/exercise/?muscles=%d&language=2&status=2", wgerURL, muscleID)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	var data models.ExerciseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	cacheSet(cacheKey, data.Results)
	return data.Results, nil
}

func GetSimilarMuscles(muscleName string) []string {
	similar := map[string][]string{
		"biceps":  {"forearms", "brachialis"},
		"triceps": {"shoulders", "chest"},
		"quads":   {"hamstrings", "glutes", "soleus"},
		"lats":    {"trapezius", "serratus anterior"},
		"abs":     {"rectus abdominis", "obliquus externus abdominis"},
	}
	if v, ok := similar[strings.ToLower(muscleName)]; ok {
		return v
	}
	return []string{"back", "core"}
}
