package apifootball

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	baseURL  = "https://v3.football.api-sports.io"
	cacheTTL = 24 * time.Hour
	dirPerm  = 0700
	filePerm = 0600
)

// ClientInterface defines the contract for the API-Football adapter.
// Tests and generate command can mock this.
type ClientInterface interface {
	SearchPlayer(name string) ([]PlayerResult, error)
	SearchTeam(name string) ([]TeamResult, error)
	GetH2H(teamA, teamB int) (*H2HStats, error)
}

// Client is the concrete API-Football HTTP adapter.
type Client struct {
	apiKey   string
	cacheDir string
	http     HTTPDoer
}

// HTTPDoer abstracts http.Client.Do for testing.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates a new API-Football client.
func NewClient(apiKey, cacheDir string, httpClient HTTPDoer) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		apiKey:   apiKey,
		cacheDir: cacheDir,
		http:     httpClient,
	}
}

// SearchPlayer looks up a player by name.
func (c *Client) SearchPlayer(name string) ([]PlayerResult, error) {
	endpoint := fmt.Sprintf("/players?search=%s&season=2024", url.QueryEscape(name))
	var resp playerSearchResponse
	if err := c.fetchJSON(endpoint, &resp); err != nil {
		return nil, fmt.Errorf("search player %q: %w", name, err)
	}
	results := make([]PlayerResult, 0, len(resp.Response))
	for _, r := range resp.Response {
		pr := PlayerResult{
			ID:    r.Player.ID,
			Name:  r.Player.Name,
			Photo: r.Player.Photo,
		}
		if len(r.Statistics) > 0 {
			stat := r.Statistics[0]
			pr.TeamName = stat.Team.Name
			pr.TeamLogo = stat.Team.Logo
			pr.Goals = stat.Goals.Total
			pr.Assists = stat.Goals.Assists
			pr.Games = stat.Games.Appearances
		}
		results = append(results, pr)
	}
	return results, nil
}

// SearchTeam looks up a team by name.
func (c *Client) SearchTeam(name string) ([]TeamResult, error) {
	endpoint := fmt.Sprintf("/teams?search=%s", url.QueryEscape(name))
	var resp teamSearchResponse
	if err := c.fetchJSON(endpoint, &resp); err != nil {
		return nil, fmt.Errorf("search team %q: %w", name, err)
	}
	results := make([]TeamResult, 0, len(resp.Response))
	for _, r := range resp.Response {
		results = append(results, TeamResult{
			ID:   r.Team.ID,
			Name: r.Team.Name,
			Logo: r.Team.Logo,
		})
	}
	return results, nil
}

// GetH2H fetches head-to-head stats between two teams.
func (c *Client) GetH2H(teamA, teamB int) (*H2HStats, error) {
	endpoint := fmt.Sprintf("/fixtures/headtohead?h2h=%d-%d&last=20", teamA, teamB)
	var resp h2hResponse
	if err := c.fetchJSON(endpoint, &resp); err != nil {
		return nil, fmt.Errorf("get h2h %d vs %d: %w", teamA, teamB, err)
	}
	stats := &H2HStats{Matches: len(resp.Response)}
	for _, fix := range resp.Response {
		homeGoals := fix.Goals.Home
		awayGoals := fix.Goals.Away
		if fix.Teams.Home.ID == teamA {
			stats.GoalsA += homeGoals
			stats.GoalsB += awayGoals
			switch {
			case homeGoals > awayGoals:
				stats.WinsA++
			case awayGoals > homeGoals:
				stats.WinsB++
			default:
				stats.Draws++
			}
		} else {
			stats.GoalsA += awayGoals
			stats.GoalsB += homeGoals
			switch {
			case awayGoals > homeGoals:
				stats.WinsA++
			case homeGoals > awayGoals:
				stats.WinsB++
			default:
				stats.Draws++
			}
		}
	}
	return stats, nil
}

// fetchJSON performs a cached HTTP GET and unmarshals the response.
func (c *Client) fetchJSON(endpoint string, dst any) error {
	cacheKey := cacheKeyFor(endpoint)
	if data, ok := c.readCache(cacheKey); ok {
		return json.Unmarshal(data, dst)
	}

	reqURL := baseURL + endpoint
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-apisports-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}

	c.writeCache(cacheKey, body)

	return json.Unmarshal(body, dst)
}

func cacheKeyFor(endpoint string) string {
	h := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(h[:8])
}

func (c *Client) cachePath(key string) string {
	return filepath.Join(c.cacheDir, key+".json")
}

func (c *Client) readCache(key string) ([]byte, bool) {
	path := c.cachePath(key)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > cacheTTL {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func (c *Client) writeCache(key string, data []byte) {
	if c.cacheDir == "" {
		return
	}
	_ = os.MkdirAll(c.cacheDir, dirPerm)
	_ = os.WriteFile(c.cachePath(key), data, filePerm)
}

// API-Football v3 response structures (internal)

type playerSearchResponse struct {
	Response []playerSearchEntry `json:"response"`
}

type playerSearchEntry struct {
	Player struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Photo string `json:"photo"`
	} `json:"player"`
	Statistics []struct {
		Team struct {
			Name string `json:"name"`
			Logo string `json:"logo"`
		} `json:"team"`
		Games struct {
			Appearances int `json:"appearences"` // API typo is real
		} `json:"games"`
		Goals struct {
			Total   int `json:"total"`
			Assists int `json:"assists"`
		} `json:"goals"`
	} `json:"statistics"`
}

type teamSearchResponse struct {
	Response []struct {
		Team struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Logo string `json:"logo"`
		} `json:"team"`
	} `json:"response"`
}

type h2hResponse struct {
	Response []h2hFixture `json:"response"`
}

type h2hFixture struct {
	Teams struct {
		Home struct {
			ID int `json:"id"`
		} `json:"home"`
		Away struct {
			ID int `json:"id"`
		} `json:"away"`
	} `json:"teams"`
	Goals struct {
		Home int `json:"home"`
		Away int `json:"away"`
	} `json:"goals"`
}
