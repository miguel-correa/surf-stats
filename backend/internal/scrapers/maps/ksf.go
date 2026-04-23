package maps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type KSFScraper struct {
	baseURL string
	client  *http.Client
}

func NewKSFScraper(baseURL string) *KSFScraper {
	return &KSFScraper{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Map struct {
	MapID    int    `json:"mapID"`
	Name     string `json:"name"`
	Tier     int    `json:"tier"`
	Created  int    `json:"created"`
	Playtime int    `json:"playtime"`
	BCount   int    `json:"b_count"`
	IsLinear bool   `json:"isLinear"`
}

type MapsResponse struct {
	Maps []Map `json:"maps"`
}

func (s *KSFScraper) FetchMaps(path string) ([]Map, error) {
	body, err := os.ReadFile(path)

	var response MapsResponse
	if err != nil {
		log.Printf("maps-scraper: fetching %s (timeout=%s)", s.baseURL, s.client.Timeout)
		resp, err := s.client.Get(s.baseURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unexpected status code")
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return response.Maps, err
		}
	}
	html := string(body)

	// Find where maps array is - extract the full {"maps":[...]} object
	re := regexp.MustCompile(`(\{\\"maps\\":\[.*?\]\})`)
	matchResult := re.FindStringSubmatch(html)

	if len(matchResult) < 2 {
		return nil, errors.New("maps array not found")
	}

	jsonStr := matchResult[1]
	// Remove all backslashes
	jsonFixed := strings.ReplaceAll(jsonStr, `\`, ``)

	if err := json.Unmarshal([]byte(jsonFixed), &response); err != nil {
		return nil, fmt.Errorf("failed to parse maps JSON: %w", err)
	}

	return response.Maps, nil
}
