package handler

import (
	"fmt"
	//"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/m4rk1sov/rbk-api/internal/domain/models"
	"github.com/m4rk1sov/rbk-api/internal/service"
	"github.com/m4rk1sov/rbk-api/pkg/util"
	"net/http"
	"strings"
)

func HandleFitness(w http.ResponseWriter, r *http.Request) {
	muscle := strings.ToLower(chi.URLParam(r, "muscle"))

	muscleID, err := service.GetMuscleID(muscle)
	if err != nil {
		util.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Muscle not found"})
		return
	}

	exercises, err := service.GetExercisesByMuscle(muscleID)
	if err != nil {
		util.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch the exercises"})
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
		"advice":         service.GetAdvice(),
		"similarMuscles": service.GetSimilarMuscles(muscle),
	}

	util.WriteJSON(w, http.StatusOK, resp)

	//w.Header().Set("Content-Type", "application/json")
	//err = json.NewEncoder(w).Encode(resp)
	//if err != nil {
	//	return
	//}
}

func ListAvailable(w http.ResponseWriter, r *http.Request) {
	//todo list available commands
	_, err := fmt.Fprintf(w, "list of all commands")
	if err != nil {
		return
	}
}
