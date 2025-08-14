package handler

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/m4rk1sov/rbk-api/internal/service"
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"github.com/m4rk1sov/rbk-api/pkg/util"
	"strings"
	
	"net/http"
)

type FitnessSaver interface {
	SaveRequest(muscle string) error
}

func HandleFitness(repo FitnessSaver, logger *jsonlog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		muscle := strings.ToLower(chi.URLParam(r, "muscle"))
		err := repo.SaveRequest(muscle)
		if err != nil {
			logger.PrintError("failed to save to database", map[string]string{"error": err.Error()})
			util.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save to database"})
			return
		}
		
		muscleID, err := service.GetMuscleID(muscle)
		if err != nil {
			logger.PrintError("muscle not found", map[string]string{"error": err.Error()})
			util.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Muscle not found"})
			return
		}
		
		exercises, err := service.GetExercisesByMuscle(muscleID)
		if err != nil || len(exercises) == 0 {
			logger.PrintError("exercises are not found", map[string]string{"error": err.Error()})
			util.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch the exercises"})
			return
		}
		
		var cleaned []map[string]interface{}
		for _, exercise := range exercises {
			if exercise.Description != "" {
				cleaned = append(cleaned, map[string]interface{}{
					"name":        exercise.Name,
					"description": util.StripHTML(exercise.Description),
					"equipment":   exercise.Equipment,
					"images":      exercise.Images,
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
}

func ListAvailable(w http.ResponseWriter, r *http.Request) {
	//todo list available commands
	_, err := fmt.Fprintf(w, "list of all commands")
	if err != nil {
		return
	}
}
