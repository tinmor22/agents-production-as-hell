package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/statcard/statcard/internal/apifootball"
	"github.com/statcard/statcard/internal/config"
	"github.com/statcard/statcard/internal/meter"
	"github.com/statcard/statcard/internal/metrics"
	"github.com/statcard/statcard/internal/parser"
	"github.com/statcard/statcard/internal/renderer"
	"github.com/statcard/statcard/templates"
)

// version is injected at build time via -ldflags.
var version = "dev"

// configPath allows tests to override the config file location.
var configPath string

// meterDir allows tests to override the meter directory.
var meterDir string

// metricsDir allows tests to override the metrics directory.
var metricsDir string

// apiClient allows tests to inject a mock.
var apiClient apifootball.ClientInterface

// demoClient is a fake ClientInterface used when --demo flag is set.
type demoClient struct{}

func (d *demoClient) SearchPlayer(name string) ([]apifootball.PlayerResult, error) {
	demos := map[string]apifootball.PlayerResult{
		"messi":     {ID: 154, Name: "Lionel Messi", TeamName: "Inter Miami", Goals: 756, Assists: 376, Games: 1041, Photo: "https://media.api-sports.io/football/players/154.png"},
		"cristiano": {ID: 874, Name: "Cristiano Ronaldo", TeamName: "Al Nassr", Goals: 901, Assists: 232, Games: 1175, Photo: "https://media.api-sports.io/football/players/874.png"},
		"riquelme":  {ID: 3, Name: "Juan Roman Riquelme", TeamName: "Boca Juniors", Goals: 194, Assists: 255, Games: 638, Photo: "https://media.api-sports.io/football/players/3.png"},
		"zidane":    {ID: 4, Name: "Zinedine Zidane", TeamName: "Real Madrid", Goals: 125, Assists: 145, Games: 679},
		"maradona":  {ID: 5, Name: "Diego Maradona", TeamName: "Napoli", Goals: 312, Assists: 181, Games: 491},
		"pele":      {ID: 6, Name: "Pele", TeamName: "Santos", Goals: 767, Assists: 101, Games: 831},
		"mbappe":    {ID: 278, Name: "Kylian Mbappe", TeamName: "Real Madrid", Goals: 342, Assists: 198, Games: 489, Photo: "https://media.api-sports.io/football/players/278.png"},
		"neymar":    {ID: 276, Name: "Neymar Jr", TeamName: "Al Hilal", Goals: 421, Assists: 310, Games: 698, Photo: "https://media.api-sports.io/football/players/276.png"},
	}
	lower := strings.ToLower(name)
	for key, p := range demos {
		if strings.Contains(lower, key) {
			return []apifootball.PlayerResult{p}, nil
		}
	}
	return []apifootball.PlayerResult{{ID: 99, Name: name, TeamName: "Demo FC", Goals: 100, Assists: 50, Games: 300}}, nil
}

func (d *demoClient) SearchTeam(name string) ([]apifootball.TeamResult, error) {
	return []apifootball.TeamResult{{ID: 1, Name: name}}, nil
}

func (d *demoClient) GetH2H(teamA, teamB int) (*apifootball.H2HStats, error) {
	return &apifootball.H2HStats{Matches: 36, WinsA: 16, WinsB: 11, Draws: 9, GoalsA: 56, GoalsB: 48}, nil
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "statcard [prompt]",
		Short:         "Genera tarjetas de estadisticas de futbol desde un prompt en espanol",
		Long:          "StatCard CLI: un binario que toma un prompt en espanol, consulta estadisticas de futbol y genera una tarjeta PNG lista para publicar.",
		Args:          cobra.ArbitraryArgs,
		RunE:          runGenerate,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.Flags().String("watermark", "", "Texto de marca de agua (sobreescribe config)")
	root.Flags().String("format", "both", "Formato: square, portrait, both")
	root.Flags().String("output-dir", ".", "Directorio de salida para las imagenes")
	root.Flags().Bool("verbose", false, "Muestra informacion de depuracion")
	root.Flags().Bool("demo", false, "Usa datos de demo (sin API key necesaria)")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newStatusCmd())

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Muestra la version de StatCard",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "statcard %s\n", version)
		},
	}
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Inicializa la configuracion de StatCard",
		RunE:  runInit,
	}
	cmd.Flags().String("api-key", "", "Clave de API de api-football.com (requerida)")
	cmd.Flags().String("watermark", "", "Texto de marca de agua (opcional)")
	_ = cmd.MarkFlagRequired("api-key")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	apiKey, _ := cmd.Flags().GetString("api-key")
	watermark, _ := cmd.Flags().GetString("watermark")

	cfg := &config.Config{
		APIKey:    apiKey,
		Watermark: watermark,
	}

	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("guardar configuracion: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Configuracion guardada exitosamente.")
	fmt.Fprintln(cmd.OutOrStdout(), "Para obtener tu API key gratis: https://rapidapi.com/api-sports/api/api-football")
	return nil
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Muestra el estado de uso diario",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	dir := resolveDir(meterDir, cfg)
	dc, err := meter.Load(dir)
	if err != nil {
		return fmt.Errorf("cargar contador: %w", err)
	}
	_ = meter.Check(dc, cfg.DailyLimit)

	mDir := resolveDir(metricsDir, cfg)
	weekAgo := time.Now().AddDate(0, 0, -7)
	weeklyCount, _ := metrics.CountSince(mDir, weekAgo)

	fmt.Fprintf(cmd.OutOrStdout(), "Plan: %s\n", cfg.Plan)
	fmt.Fprintf(cmd.OutOrStdout(), "Tarjetas hoy: %d/%d\n", dc.Count, cfg.DailyLimit)
	fmt.Fprintf(cmd.OutOrStdout(), "Restantes: %d\n", cfg.DailyLimit-dc.Count)
	fmt.Fprintf(cmd.OutOrStdout(), "Esta semana: %d\n", weeklyCount)
	if cfg.Watermark != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Marca de agua: %s\n", cfg.Watermark)
	}
	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	start := time.Now()
	prompt := strings.Join(args, " ")
	verbose, _ := cmd.Flags().GetBool("verbose")
	demo, _ := cmd.Flags().GetBool("demo")

	// 1. Parse prompt
	pq, err := parser.Parse(prompt)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "[debug] ParsedQuery: A=%q B=%q ctx=%q\n", pq.EntityA, pq.EntityB, pq.Context)
	}

	// 2. Resolve client, meter, and config
	var client apifootball.ClientInterface
	var dc *meter.DailyCounter
	var statDir, metricsBaseDir string
	var dailyLimit int
	watermark, _ := cmd.Flags().GetString("watermark")

	if demo {
		fmt.Fprintln(cmd.ErrOrStderr(), "[demo] Usando datos de demostracion — no se requiere API key")
		client = &demoClient{}
		dailyLimit = 9999
		dc = &meter.DailyCounter{Date: time.Now().Format("2006-01-02"), Count: 0}
	} else {
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		statDir = resolveDir(meterDir, cfg)
		metricsBaseDir = resolveDir(metricsDir, cfg)
		dailyLimit = cfg.DailyLimit
		if watermark == "" {
			watermark = cfg.Watermark
		}

		dc, err = meter.Load(statDir)
		if err != nil {
			return fmt.Errorf("cargar contador: %w", err)
		}
		if err := meter.Check(dc, cfg.DailyLimit); err != nil {
			return err
		}

		client = apiClient
		if client == nil {
			cacheDir := filepath.Join(statDir, "cache")
			client = apifootball.NewClient(cfg.APIKey, cacheDir, nil)
		}
	}

	// 3. Search players
	playersA, err := client.SearchPlayer(pq.EntityA)
	if err != nil {
		return fmt.Errorf("buscar %q: %w", pq.EntityA, err)
	}
	if len(playersA) == 0 {
		return fmt.Errorf("no se encontro jugador: %q", pq.EntityA)
	}

	playersB, err := client.SearchPlayer(pq.EntityB)
	if err != nil {
		return fmt.Errorf("buscar %q: %w", pq.EntityB, err)
	}
	if len(playersB) == 0 {
		return fmt.Errorf("no se encontro jugador: %q", pq.EntityB)
	}

	playerA := playersA[0]
	playerB := playersB[0]

	if verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "[debug] Player A: %s (%s)\n", playerA.Name, playerA.TeamName)
		fmt.Fprintf(cmd.ErrOrStderr(), "[debug] Player B: %s (%s)\n", playerB.Name, playerB.TeamName)
	}

	// 4. Download player photos (best-effort, won't fail if unavailable)
	photoA := downloadPhoto(playerA.Photo)
	photoB := downloadPhoto(playerB.Photo)

	// 5. Build CardData
	cardData := renderer.CardData{
		Title:       fmt.Sprintf("%s vs %s", strings.ToUpper(playerA.Name), strings.ToUpper(playerB.Name)),
		Subtitle:    buildSubtitle(pq, playerA, playerB),
		EntityA:     renderer.EntityInfo{Name: playerA.Name, AccentColor: "#00A3E0", PhotoData: photoA},
		EntityB:     renderer.EntityInfo{Name: playerB.Name, AccentColor: "#E4002B", PhotoData: photoB},
		Stats:       buildPlayerStats(playerA, playerB),
		Watermark:   watermark,
		GeneratedAt: time.Now().Format("2006-01-02 15:04"),
	}

	// 5. Render card
	r, err := renderer.New(templates.FS)
	if err != nil {
		return fmt.Errorf("inicializar renderer: %w", err)
	}

	formatFlag, _ := cmd.Flags().GetString("format")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	formats := resolveFormats(formatFlag)

	paths, err := r.RenderCard(cardData, formats, outputDir)
	if err != nil {
		return fmt.Errorf("renderizar tarjeta: %w", err)
	}

	// 6. Increment meter (skip in demo mode)
	if !demo {
		meter.Increment(dc)
		if statDir != "" {
			_ = meter.Save(dc, statDir)
		}
		_ = metrics.Append(metricsBaseDir, metrics.Entry{
			EntityA: playerA.Name,
			EntityB: playerB.Name,
			Formats: len(formats),
			Elapsed: time.Since(start).Seconds(),
			Success: true,
		})
	}

	// 7. Print results
	fmt.Fprintf(cmd.OutOrStdout(), "Tarjetas generadas en %.1fs:\n", time.Since(start).Seconds())
	for _, p := range paths {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", p)
	}
	if !demo {
		fmt.Fprintf(cmd.OutOrStdout(), "Uso: %d/%d tarjetas hoy\n", dc.Count, dailyLimit)
	}

	return nil
}

func buildSubtitle(pq *parser.ParsedQuery, a, b apifootball.PlayerResult) string {
	parts := []string{a.TeamName, "vs", b.TeamName}
	if pq.Context != "" {
		parts = append(parts, "-", pq.Context)
	}
	return strings.Join(parts, " ")
}

func buildPlayerStats(a, b apifootball.PlayerResult) []renderer.StatRow {
	winnerGoals := "none"
	if a.Goals > b.Goals {
		winnerGoals = "a"
	} else if b.Goals > a.Goals {
		winnerGoals = "b"
	}
	winnerAssists := "none"
	if a.Assists > b.Assists {
		winnerAssists = "a"
	} else if b.Assists > a.Assists {
		winnerAssists = "b"
	}
	_ = winnerGoals
	_ = winnerAssists
	return []renderer.StatRow{
		{Label: "Goles", ValueA: fmt.Sprintf("%d", a.Goals), ValueB: fmt.Sprintf("%d", b.Goals)},
		{Label: "Asistencias", ValueA: fmt.Sprintf("%d", a.Assists), ValueB: fmt.Sprintf("%d", b.Assists)},
		{Label: "Partidos", ValueA: fmt.Sprintf("%d", a.Games), ValueB: fmt.Sprintf("%d", b.Games)},
	}
}

func resolveFormats(flag string) []string {
	switch strings.ToLower(flag) {
	case "square":
		return []string{renderer.FormatSquare}
	case "portrait":
		return []string{renderer.FormatPortrait}
	default:
		return []string{renderer.FormatSquare, renderer.FormatPortrait}
	}
}

// downloadPhoto fetches a photo URL and returns the bytes, or nil on any error.
func downloadPhoto(url string) []byte {
	if url == "" {
		return nil
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return data
}

func resolveDir(override string, cfg *config.Config) string {
	if override != "" {
		return override
	}
	dir, err := config.DefaultDir()
	if err != nil {
		return ".statcard"
	}
	return dir
}

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
