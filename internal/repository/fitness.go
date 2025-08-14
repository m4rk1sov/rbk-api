package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m4rk1sov/rbk-api/internal/domain/models"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type WgerClient struct {
	httpClient *http.Client
	baseURL    string
	language   int
	userAgent  string
}

func NewWgerClient(httpClient *http.Client, baseURL string, language int, userAgent string) *WgerClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 60 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	}
	if baseURL == "" {
		baseURL = "https://wger.de/api/v2"
	}
	if language == 0 {
		language = 2
	}
	if userAgent == "" {
		userAgent = "rbk-api/1.0 (+https://github.com/m4rk1sov/rbk-api)"
	}
	return &WgerClient{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
		language:   language,
		userAgent:  userAgent,
	}
}

type wgerExercise struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Category         int    `json:"category"`
	Muscles          []int  `json:"muscles"`
	MusclesSecondary []int  `json:"muscles_secondary"`
	Equipment        []int  `json:"equipment"`
}

type wgerPagedResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []wgerExercise `json:"results"`
}

// FetchExercises fetches from primary and secondary muscles and merges results (deduplicated by ID).
func (c *WgerClient) FetchExercises(ctx context.Context, muscles []int, limit int) ([]models.Exercise, error) {
	if len(muscles) == 0 {
		return nil, errors.New("no muscles provided")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Helper to call endpoint with given query
	call := func(param string, muscleIDs []int) ([]wgerExercise, error) {
		u, err := url.Parse(c.baseURL + "/exercise/")
		if err != nil {
			return nil, err
		}

		q := u.Query()
		q.Set("language", strconv.Itoa(c.language))
		q.Set("limit", strconv.Itoa(limit))
		// wger supports filter by 'muscles' and 'muscles_secondary'
		q.Set(param, intsToCSV(muscleIDs))
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			if closeErr := Body.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}(resp.Body)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("wger returned %d", resp.StatusCode)
		}

		var pr wgerPagedResponse
		if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
			return nil, err
		}
		return pr.Results, nil
	}

	primary, err := call("muscles", muscles)
	if err != nil {
		return nil, err
	}
	secondary, err := call("muscles_secondary", muscles)
	if err != nil {
		// secondary may be empty
		return nil, err
	}

	merged := make(map[int]wgerExercise, len(primary)+len(secondary))
	for _, e := range primary {
		merged[e.ID] = e
	}
	for _, e := range secondary {
		merged[e.ID] = e
	}

	out := make([]models.Exercise, 0, len(merged))
	for _, e := range merged {
		out = append(out, models.Exercise{
			ID:               e.ID,
			Name:             e.Name,
			Description:      e.Description,
			Category:         e.Category,
			Muscles:          e.Muscles,
			MusclesSecondary: e.MusclesSecondary,
			Equipment:        e.Equipment,
		})
	}
	return out, nil
}

func intsToCSV(v []int) string {
	if len(v) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i, n := range v {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Itoa(n))
	}
	return sb.String()
}
