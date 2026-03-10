package apifootball

// PlayerResult holds key player data from API-Football.
type PlayerResult struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	TeamName string `json:"team_name"`
	TeamLogo string `json:"team_logo"`
	Photo    string `json:"photo"`
	Goals    int    `json:"goals"`
	Assists  int    `json:"assists"`
	Games    int    `json:"games"`
}

// TeamResult holds key team data from API-Football.
type TeamResult struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

// H2HStats holds head-to-head statistics.
type H2HStats struct {
	PlayerA PlayerResult `json:"player_a"`
	PlayerB PlayerResult `json:"player_b"`
	Matches int          `json:"matches"`
	WinsA   int          `json:"wins_a"`
	WinsB   int          `json:"wins_b"`
	Draws   int          `json:"draws"`
	GoalsA  int          `json:"goals_a"`
	GoalsB  int          `json:"goals_b"`
}
