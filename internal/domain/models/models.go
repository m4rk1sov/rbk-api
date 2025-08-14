package models

type Muscle struct {
	ID   int    `json:"id"`
	Name string `json:"name_en"`
}

type MuscleResponse struct {
	Results []Muscle `json:"results"`
}

type Equipment struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExerciseInfo struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Equipment   []Equipment `json:"equipment"`
	Images      []struct {
		Image string `json:"image"`
	} `json:"images"`
}

type ExerciseInfoResponse struct {
	Results []ExerciseInfo `json:"results"`
}

type ExerciseOut struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AdviceSlip struct {
	ID     int    `json:"id"`
	Advice string `json:"advice"`
}

type AdviceResponse struct {
	Slip AdviceSlip `json:"slip"`
}
