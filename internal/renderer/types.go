package renderer

// CardData holds all data needed to render a stat card.
type CardData struct {
	Title       string
	Subtitle    string
	EntityA     EntityInfo
	EntityB     EntityInfo
	Stats       []StatRow
	Watermark   string
	GeneratedAt string
}

// EntityInfo describes one side of a head-to-head comparison.
type EntityInfo struct {
	Name        string
	AccentColor string
}

// StatRow is a single stat comparison line on the card.
type StatRow struct {
	Label  string
	ValueA string
	ValueB string
}
