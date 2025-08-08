package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/m4rk1sov/rbk-api/internal/models"
	"github.com/m4rk1sov/rbk-api/internal/service"
	"github.com/m4rk1sov/rbk-api/internal/util"
	"net/http"
	"strings"
)

func HandleFitness(w http.ResponseWriter, r *http.Request) {
	muscle := strings.ToLower(chi.URLParam(r, "muscle"))
	
	muscleID, err := service.GetMuscleID(muscle)
	if err != nil {
		http.Error(w, "Muscle not found", http.StatusNotFound)
		return
	}
	
	exercises, err := service.GetExercisesByMuscle(muscleID)
	if err != nil {
		http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
		return
	}
	
	var cleaned []models.ExerciseOut
	for _, exercise := range exercises {
		if exercise.Description != "" {
			cleaned = append(cleaned, models.ExerciseOut{
				Name:        exercise.Name,
				Description: util.StripHTML(exercise.Description),
			})
		}
	}
	
	resp := map[string]interface{}{
		"muscle":         muscle,
		"exercises":      cleaned,
		"advice":         "Stay hydrated!",
		"similarMuscles": []string{"forearms", "brachialis"},
	}
	
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}
