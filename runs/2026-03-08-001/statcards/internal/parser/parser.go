package parser

import (
	"errors"
	"regexp"
	"strings"
)

// QueryType constants for parsed prompts.
const (
	QueryTypeH2H = "h2h" // head-to-head comparison
)

// ErrNoSeparator is returned when the prompt has no vs/contra separator.
var ErrNoSeparator = errors.New("no se encontro un separador (vs, contra). Ejemplo: 'Messi vs Cristiano'")

// ParsedQuery holds the structured result of parsing a Spanish prompt.
type ParsedQuery struct {
	EntityA   string
	EntityB   string
	Context   string
	QueryType string
}

// separatorPattern matches "vs", "VS", "contra", "Contra" etc.
// The context keyword "en" is optional after the second entity.
var separatorPattern = regexp.MustCompile(`(?i)\s+(vs\.?|contra)\s+`)
var contextPattern = regexp.MustCompile(`(?i)\s+en\s+`)

// Parse converts a Spanish football prompt into a ParsedQuery.
// It expects a separator (vs/contra) between two entity names.
// An optional "en <context>" clause can follow the second entity.
func Parse(prompt string) (*ParsedQuery, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, ErrNoSeparator
	}

	loc := separatorPattern.FindStringIndex(prompt)
	if loc == nil {
		return nil, ErrNoSeparator
	}

	entityA := strings.TrimSpace(prompt[:loc[0]])
	rest := strings.TrimSpace(prompt[loc[1]:])

	if entityA == "" || rest == "" {
		return nil, ErrNoSeparator
	}

	var entityB, context string

	ctxLoc := contextPattern.FindStringIndex(rest)
	if ctxLoc != nil {
		entityB = strings.TrimSpace(rest[:ctxLoc[0]])
		context = strings.TrimSpace(rest[ctxLoc[1]:])
	} else {
		entityB = rest
	}

	if entityB == "" {
		return nil, ErrNoSeparator
	}

	return &ParsedQuery{
		EntityA:   entityA,
		EntityB:   entityB,
		Context:   context,
		QueryType: QueryTypeH2H,
	}, nil
}
