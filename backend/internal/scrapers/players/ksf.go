package players

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type KSFScraper struct {
	baseURL string
	client  *http.Client
}

func NewKSFScraper() *KSFScraper {
	return &KSFScraper{
		baseURL: "https://ksf.surf/api/players",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type completionsResponse struct {
	MapID string `json:"mapID"`
	Zones []struct {
		ZoneID     int `json:"zoneId"`
		TotalRanks int `json:"totalRanks"`
	} `json:"records"`
}

// Gets the total number of completions via the player record API.
// Used when scraping the maps, as the completions are not included in the maps page.
// Returns mapID, completions, err
func (s *KSFScraper) FetchMapCompletionsFromPlayerRecord(steamID string, mapName string) (int, int, error) {

	url := s.baseURL + "/" + steamID + "/records/map/" + mapName + "?game=css&mode=0"
	resp, err := s.client.Get(url)
	if err != nil {
		return -1, -1, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return -1, -1, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, -1, err
	}

	var response completionsResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return -1, -1, fmt.Errorf("failed to parse map records JSON: %w", err)
	}
	mapID, err := strconv.Atoi(response.MapID)
	if err != nil {
		return -1, -1, fmt.Errorf("failed parsing mapID to int: %v", err)
	}
	for _, zone := range response.Zones {
		if zone.ZoneID == 0 {
			return mapID, zone.TotalRanks, nil
		}
	}

	return -1, -1, fmt.Errorf("failed to find completions for given map")
}
