package service

import (
	"encoding/json"
	"github.com/m4rk1sov/rbk-api/internal/domain/models"
	"io"
	"net/http"
	"time"
)

const adviceURL = "https://api.adviceslip.com/advice"

var httpClientAdvice = &http.Client{Timeout: 10 * time.Second}

const fail = "no advice for today"

func GetAdvice() string {
	resp, err := httpClientAdvice.Get(adviceURL)
	if err != nil {
		return fail
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			return
		}
	}(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fail
	}
	
	var data models.AdviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fail
	}
	if data.Slip.Advice == "" {
		return fail
	}
	
	return data.Slip.Advice
}
