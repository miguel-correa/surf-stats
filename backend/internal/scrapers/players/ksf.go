package players

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"surfstats/internal/models"
	"time"
)

type KSFScraper struct {
	baseURL string
	client  *http.Client
}

func NewKSFScraper() *KSFScraper {
	return newKSFScraper("https://ksf.surf/api/players", &http.Client{
		Timeout: 10 * time.Second,
	})
}

func newKSFScraper(baseURL string, client *http.Client) *KSFScraper {
	return &KSFScraper{
		baseURL: baseURL,
		client:  client,
	}
}

type mapRecordsResponse struct {
	MapID     string `json:"mapID"`
	BasicInfo struct {
		SteamID  string `json:"steamID"`
		PlayerID int    `json:"playerID"`
	} `json:"basicInfo"`
	Zones []struct {
		ZoneID      int      `json:"zoneId"`
		StageID     string   `json:"stageID"`
		SurfTime    *float64 `json:"surfTime"`
		Rank        int      `json:"rank"`
		TotalRanks  int      `json:"totalRanks"`
		DateSet     string   `json:"date_set"`
		Completions *int     `json:"completions"`
		Group       *int     `json:"group"`
	} `json:"records"`
}

// Gets the total number of completions via the player record API.
// Used when scraping the maps, as the completions are not included in the maps page.
// Returns mapID, completions, err
func (s *KSFScraper) FetchMapCompletionsFromPlayerRecord(steamID string, mapName string) (int, int, error) {
	response, err := s.fetchMapRecords(steamID, mapName)
	if err != nil {
		return -1, -1, err
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

func (s *KSFScraper) FetchMainMapRecord(steamID string, mapName string) (*models.PlayerMapRecord, error) {
	response, err := s.fetchMapRecords(steamID, mapName)
	if err != nil {
		return nil, err
	}

	mapID, err := strconv.Atoi(response.MapID)
	if err != nil {
		return nil, fmt.Errorf("failed parsing mapID to int: %v", err)
	}

	for _, zone := range response.Zones {
		if zone.ZoneID != 0 || zone.SurfTime == nil {
			continue
		}

		record := models.PlayerMapRecord{
			SteamID:     response.BasicInfo.SteamID,
			PlayerID:    response.BasicInfo.PlayerID,
			KSFMapID:    mapID,
			SurfTimeMS:  int(math.Round(*zone.SurfTime * 1000)),
			Rank:        zone.Rank,
			TotalRanks:  zone.TotalRanks,
			Completions: zone.Completions,
			GroupTier:   zone.Group,
		}

		dateSet, err := parseAPITime(zone.DateSet)
		if err != nil {
			return nil, err
		}
		record.DateSet = dateSet

		return &record, nil
	}

	return nil, nil
}

func (s *KSFScraper) fetchMapRecords(steamID string, mapName string) (mapRecordsResponse, error) {
	endpoint, err := url.Parse(s.baseURL)
	if err != nil {
		return mapRecordsResponse{}, fmt.Errorf("invalid base URL %q: %w", s.baseURL, err)
	}

	endpoint.Path = path.Join(endpoint.Path, steamID, "records", "map", mapName)
	query := endpoint.Query()
	query.Set("game", "css")
	query.Set("mode", "0")
	endpoint.RawQuery = query.Encode()

	resp, err := s.client.Get(endpoint.String())
	if err != nil {
		return mapRecordsResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return mapRecordsResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mapRecordsResponse{}, err
	}

	var response mapRecordsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return mapRecordsResponse{}, fmt.Errorf("failed to parse map records JSON: %w", err)
	}

	return response, nil
}

func parseAPITime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "1970-01-01T00:00:00.000Z" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("parse API time %q: %w", value, err)
	}
	return &parsed, nil
}
