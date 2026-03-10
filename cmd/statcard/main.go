package main

import (
	"fmt"
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
// Empty string means use the default (~/.statcard/config.json).
var configPath string

// meterDir allows tests to override the meter directory.
var meterDir string

// metricsDir allows tests to override the metrics directory.
var metricsDir string

// apiClient allows tests to inject a mock.
var apiClient apifootball.ClientInterface

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "statcard [prompt]",
		Short: "Genera tarjetas de estadisticas de futbol desde un prompt en espanol",
		Long:  "StatCard CLI: un binario que toma un prompt en espanol, consulta estadisticas de futbol y genera una tarjeta PNG lista para publicar.",
		Args:  cobra.ArbitraryArgs,
		RunE:  runGenerate,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.Flags().String("watermark", "", "Texto de marca de agua (sobreescribe config)")
	root.Flags().String("format", "both", "Formato: square, portrait, both")
	root.Flags().String("output-dir", ".", "Directorio de salida para las imagenes")
	root.Flags().Bool("verbose", false, "Muestra informacion de depuracion")

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
	// Reset if date changed so we show current day
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

	// 1. Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	dir := resolveDir(meterDir, cfg)

	// 2. Check daily limit
	dc, err := meter.Load(dir)
	if err != nil {
		return fmt.Errorf("cargar contador: %w", err)
	}
	if err := meter.Check(dc, cfg.DailyLimit); err != nil {
		return err
	}

	// 3. Parse prompt
	pq, err := parser.Parse(prompt)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "[debug] ParsedQuery: A=%q B=%q ctx=%q\n", pq.EntityA, pq.EntityB, pq.Context)
	}

	// 4. Search players
	client := apiClient
	if client == nil {
		cacheDir := filepath.Join(dir, "cache")
		client = apifootball.NewClient(cfg.APIKey, cacheDir, nil)
	}

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

	// 5. Build CardData
	cardData := renderer.CardData{
		Title:    fmt.Sprintf("%s vs %s", playerA.Name, playerB.Name),
		Subtitle: buildSubtitle(pq, playerA, playerB),
		EntityA: renderer.EntityInfo{
			Name:        playerA.Name,
			AccentColor: "#00A3E0",
		},
		EntityB: renderer.EntityInfo{
			Name:        playerB.Name,
			AccentColor: "#E4002B",
		},
		Stats:       buildPlayerStats(playerA, playerB),
		GeneratedAt: time.Now().Format("2006-01-02 15:04"),
	}

	// Apply watermark
	wm, _ := cmd.Flags().GetString("watermark")
	if wm == "" {
		wm = cfg.Watermark
	}
	cardData.Watermark = wm

	// 6. Render card
	r, err := renderer.New(templates.FS)
	if err != nil {
		return fmt.Errorf("inicializar renderer: %w", err)
	}

	formatFlag, _ := cmd.Flags().GetString("format")
	formats := resolveFormats(formatFlag)

	outputDir, _ := cmd.Flags().GetString("output-dir")

	paths, err := r.RenderCard(cardData, formats, outputDir)
	if err != nil {
		return fmt.Errorf("renderizar tarjeta: %w", err)
	}

	// 7. Increment meter + save
	meter.Increment(dc)
	if err := meter.Save(dc, dir); err != nil {
		return fmt.Errorf("guardar contador: %w", err)
	}

	// 8. Append metrics
	mDir := resolveDir(metricsDir, cfg)
	_ = metrics.Append(mDir, metrics.Entry{
		EntityA: playerA.Name,
		EntityB: playerB.Name,
		Formats: len(formats),
		Elapsed: time.Since(start).Seconds(),
		Success: true,
	})

	// 9. Print results
	elapsed := time.Since(start)
	fmt.Fprintf(cmd.OutOrStdout(), "Tarjetas generadas en %.1fs:\n", elapsed.Seconds())
	for _, p := range paths {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", p)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Uso: %d/%d tarjetas hoy\n", dc.Count, cfg.DailyLimit)

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
