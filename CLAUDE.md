# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Go project aimed at rapid production deployment of agents.

## Commands

```bash
go build ./...        # Build all packages
go test ./...         # Run all tests
go test ./pkg/... -run TestName  # Run a single test
go vet ./...          # Static analysis
go mod tidy           # Tidy dependencies
```
