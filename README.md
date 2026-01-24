## Crawler

Console crawler for website analysis. It crawls a site, validates links and
static assets, collects SEO metrics (title, meta description, h1), and produces
a JSON report per page.

### Hexlet tests and linter status:
[![Actions Status](https://github.com/aksioto/go-project-316/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/aksioto/go-project-316/actions)
[![CI](https://github.com/aksioto/go-project-316/actions/workflows/ci.yml/badge.svg)](https://github.com/aksioto/go-project-316/actions/workflows/ci.yml)

## Installation

```bash
go mod download
```

## Commands

### Build

```bash
make build
```

Builds the `bin/hexlet-go-crawler` binary.

### Test

```bash
make test
```

Runs all project tests.

### Run

```bash
make run URL=https://example.com
```

If no URL is provided, it prints help:

```bash
make run
```

### Direct run

```bash
go run ./cmd/hexlet-go-crawler https://example.com
```

### Help

```bash
go run ./cmd/hexlet-go-crawler --help
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `--depth` | Crawl depth | 10 |
| `--retries` | Number of retries for failed requests | 1 |
| `--delay` | Delay between requests | 0s |
| `--timeout` | Request timeout | 15s |
| `--rps` | Requests per second limit | 0 |
| `--user-agent` | Custom User-Agent | - |
| `--workers` | Number of workers | 4 |