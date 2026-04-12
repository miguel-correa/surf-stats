package players

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchMainMapRecordParsesZoneZeroPB(t *testing.T) {
	scraper := newKSFScraper("https://example.test/api/players", &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return jsonResponse(`{
				"mapID": "285",
				"basicInfo": {
					"steamID": "STEAM_0:1:75949009",
					"playerID": 949217
				},
				"records": [
					{
						"zoneId": 0,
						"stageID": "1546",
						"surfTime": 105.185569,
						"rank": 142,
						"totalRanks": 6614,
						"date_set": "2026-04-10T19:13:42.000Z",
						"completions": 7,
						"group": 1
					},
					{
						"zoneId": 1,
						"stageID": "1547",
						"surfTime": 5.512452,
						"rank": 338,
						"totalRanks": 19744
					}
				]
			}`)
		}),
	})

	record, err := scraper.FetchMainMapRecord("STEAM_0:1:75949009", "surf_kz_protraining")
	if err != nil {
		t.Fatalf("FetchMainMapRecord returned error: %v", err)
	}
	if record == nil {
		t.Fatal("FetchMainMapRecord returned nil record")
	}

	if record.SteamID != "STEAM_0:1:75949009" {
		t.Fatalf("SteamID = %q, want %q", record.SteamID, "STEAM_0:1:75949009")
	}
	if record.PlayerID != 949217 {
		t.Fatalf("PlayerID = %d, want %d", record.PlayerID, 949217)
	}
	if record.KSFMapID != 285 {
		t.Fatalf("KSFMapID = %d, want %d", record.KSFMapID, 285)
	}
	if record.SurfTimeMS != 105186 {
		t.Fatalf("SurfTimeMS = %d, want %d", record.SurfTimeMS, 105186)
	}
	if record.Rank != 142 {
		t.Fatalf("Rank = %d, want %d", record.Rank, 142)
	}
	if record.TotalRanks != 6614 {
		t.Fatalf("TotalRanks = %d, want %d", record.TotalRanks, 6614)
	}
	if record.Completions == nil || *record.Completions != 7 {
		t.Fatalf("Completions = %v, want 7", record.Completions)
	}
	if record.GroupTier == nil || *record.GroupTier != 1 {
		t.Fatalf("GroupTier = %v, want 1", record.GroupTier)
	}
	if record.DateSet == nil {
		t.Fatal("DateSet = nil, want non-nil")
	}
}

func TestFetchMainMapRecordReturnsNilWhenZoneZeroMissingOrUnset(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{
			name: "missing zone zero",
			body: `{
				"mapID": "285",
				"basicInfo": {"steamID": "STEAM_0:1:75949009", "playerID": 949217},
				"records": [{"zoneId": 1, "surfTime": 5.5, "rank": 1, "totalRanks": 10}]
			}`,
		},
		{
			name: "zone zero null surf time",
			body: `{
				"mapID": "285",
				"basicInfo": {"steamID": "STEAM_0:1:75949009", "playerID": 949217},
				"records": [{"zoneId": 0, "surfTime": null, "rank": 1, "totalRanks": 10}]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scraper := newKSFScraper("https://example.test/api/players", &http.Client{
				Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
					return jsonResponse(tc.body)
				}),
			})

			record, err := scraper.FetchMainMapRecord("STEAM_0:1:75949009", "surf_kz_protraining")
			if err != nil {
				t.Fatalf("FetchMainMapRecord returned error: %v", err)
			}
			if record != nil {
				t.Fatalf("FetchMainMapRecord returned %+v, want nil", *record)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}
