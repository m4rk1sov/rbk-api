package models

type Exercise struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Category         int    `json:"category"`
	Muscles          []int  `json:"muscles"`
	MusclesSecondary []int  `json:"muscles_secondary"`
	Equipment        []int  `json:"equipment"`
}

type ExercisesResponse struct {
	Muscle         string     `json:"muscle"`
	Exercises      []Exercise `json:"exercises"`
	SimilarMuscles []string   `json:"similar_muscles,omitempty"`
	Advice         string     `json:"advice,omitempty"`
}

type AdviceSlip struct {
	ID     int    `json:"id"`
	Advice string `json:"advice"`
}

type AdviceResponse struct {
	Slip AdviceSlip `json:"slip"`
}
