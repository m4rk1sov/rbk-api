package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/m4rk1sov/rbk-api/internal/domain/models"
	"github.com/m4rk1sov/rbk-api/internal/repository"
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FitnessService struct {
	client         *repository.WgerClient
	logger         *jsonlog.Logger
	similarMuscles map[string][]string
	cacheMu        sync.RWMutex
	cache          map[string]cacheItem
	ttl            time.Duration
}

type cacheItem struct {
	expiresAt time.Time
	data      []models.Exercise
}

func NewFitnessService(client *repository.WgerClient, logger *jsonlog.Logger, similarFile string) *FitnessService {
	fs := &FitnessService{
		client: client,
		logger: logger,
		cache:  make(map[string]cacheItem),
		ttl:    5 * time.Minute,
	}
	fs.similarMuscles = fs.loadSimilar(similarFile)
	return fs
}

func (s *FitnessService) loadSimilar(path string) map[string][]string {
	if path == "" {
		path = "./similar_muscles.json"
	}
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		// fallback defaults for a few common groups
		return map[string][]string{
			"chest":      {"triceps", "shoulders"},
			"back":       {"biceps", "forearms"},
			"biceps":     {"forearms", "back"},
			"triceps":    {"chest", "shoulders"},
			"legs":       {"glutes", "calves"},
			"shoulders":  {"triceps", "chest"},
			"abs":        {"obliques", "lower back"},
			"hamstrings": {"glutes", "lower back"},
			"quads":      {"glutes", "hamstrings"},
		}
	}
	defer func(f *os.File) {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}(f)
	var m map[string][]string
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return map[string][]string{}
	}
	return m
}

// list of muscles
var muscleNameToIDs = map[string][]int{
	"biceps":     {1},
	"shoulders":  {2},
	"chest":      {4},
	"triceps":    {5},
	"abs":        {6},
	"calves":     {7},
	"glutes":     {8},
	"quadriceps": {10},
	"quads":      {10},
	"hamstrings": {11},
	"lats":       {12},
	"lower back": {13},
	"trapezius":  {14},
	"forearms":   {9},
	"neck":       {3},
	"back":       {12, 13, 14}, // aggregate
}

// GetExercisesByMuscle fetches exercises and returns a domain response with advice and similar groups.
func (s *FitnessService) GetExercisesByMuscle(ctx context.Context, muscle string, limit int) (models.ExercisesResponse, error) {
	muscleKey := strings.ToLower(strings.TrimSpace(muscle))
	ids, ok := muscleNameToIDs[muscleKey]
	if !ok {
		// allow passing raw numeric id(s) comma-separated
		if csvIDs, err := parseIDsCSV(muscleKey); err == nil && len(csvIDs) > 0 {
			ids = csvIDs
		} else {
			return models.ExercisesResponse{}, errors.New("unknown muscle group; try one of: chest, back, biceps, triceps, shoulders, quads, hamstrings, calves, abs")
		}
	}
	
	cacheKey := cacheKeyFor(muscleKey, limit)
	if data, ok := s.getCache(cacheKey); ok {
		return models.ExercisesResponse{
			Muscle:         muscleKey,
			Exercises:      data,
			SimilarMuscles: s.similarMuscles[muscleKey],
			Advice:         s.makeAdvice(data),
		}, nil
	}
	
	data, err := s.client.FetchExercises(ctx, ids, limit)
	if err != nil {
		return models.ExercisesResponse{}, err
	}
	s.setCache(cacheKey, data)
	
	return models.ExercisesResponse{
		Muscle:         muscleKey,
		Exercises:      data,
		SimilarMuscles: s.similarMuscles[muscleKey],
		Advice:         s.makeAdvice(data),
	}, nil
}

func (s *FitnessService) makeAdvice(exs []models.Exercise) string {
	// Very simple heuristic advices:
	if len(exs) == 0 {
		return "Try broad compound movements and re-check your filters."
	}
	// If many exercises hit secondary muscles, suggest warm-up/isolation
	secondaryHeavy := 0
	for _, e := range exs {
		if len(e.MusclesSecondary) > 0 {
			secondaryHeavy++
		}
	}
	if secondaryHeavy > len(exs)/2 {
		return "Include specific warm-up sets and isolation moves before compounds."
	}
	return "Balance compounds with accessory work; keep proper form."
}

func cacheKeyFor(muscle string, limit int) string {
	return muscle + ":" + strconv.Itoa(limit)
}

func (s *FitnessService) getCache(key string) ([]models.Exercise, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	item, ok := s.cache[key]
	if !ok || time.Now().After(item.expiresAt) {
		return nil, false
	}
	return item.data, true
}

func (s *FitnessService) setCache(key string, data []models.Exercise) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache[key] = cacheItem{
		expiresAt: time.Now().Add(s.ttl),
		data:      data,
	}
}

func parseIDsCSV(s string) ([]int, error) {
	if s == "" {
		return nil, errors.New("empty")
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func GetAvailableMuscles() []string {
	names := make([]string, 0, len(muscleNameToIDs))
	for name := range muscleNameToIDs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
