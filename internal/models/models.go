package models

type Muscle struct {
	ID   int    `json:"id"`
	Name string `json:"name_en"`
}

type MuscleResponse struct {
	Results []Muscle `json:"results"`
}

type Exercise struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ExerciseResponse struct {
	Results []Exercise `json:"results"`
}

type ExerciseOut struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
