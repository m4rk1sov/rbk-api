package service

import (
	"encoding/json"
	"fmt"
	"github.com/m4rk1sov/rbk-api/internal/domain/models"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://wger.de/api/v2"

func GetMuscleID(muscleName string) (int, error) {
	url := baseURL + "/muscle/"
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	var data models.MuscleResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	for _, m := range data.Results {
		if strings.EqualFold(m.Name, muscleName) {
			return m.ID, nil
		}
	}

	return 0, fmt.Errorf("muscle not found")
}

func GetExercisesByMuscle(muscleID int) ([]models.Exercise, error) {
	url := fmt.Sprintf("%s/exercise/?muscles=%d&language=2&status=2", baseURL, muscleID)
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

	return data.Results, nil
}
