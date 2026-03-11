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

## Depth Parameter

The `--depth` parameter controls how many levels of links the crawler will follow:

- **depth=1**: Only the start page is analyzed (no links are followed)
- **depth=2**: Start page + pages linked directly from it
- **depth=N**: Pages up to N-1 clicks away from the start page

### Example

```text
Start Page (depth 0)
├── /about (depth 1)
│   └── /about/team (depth 2)
└── /contact (depth 1)
```

With `--depth=2`, all pages above will be crawled.  
With `--depth=1`, only the start page will be crawled.

### Notes

- Only **internal links** (same domain) are followed
- External links are checked for broken status but not crawled
- Duplicate URLs are automatically deduplicated (e.g., `/page` and `/page/index.html` are treated as the same page)

## Rate Limiting

The crawler supports two ways to control request frequency:

### Using `--delay`

Sets a fixed delay between consecutive HTTP requests:

```bash
bin/hexlet-go-crawler --delay=200ms https://example.com
```

### Using `--rps`

Sets the maximum requests per second. When specified, `--rps` **overrides** `--delay`:

```bash
bin/hexlet-go-crawler --rps=5 https://example.com
```

This limits the crawler to 5 requests per second (200ms between requests).

### Behavior

| Parameter | Effect |
|-----------|--------|
| Neither set | No rate limiting, maximum speed |
| `--delay=200ms` | 200ms pause between each request |
| `--rps=10` | ~100ms between requests (10 req/sec) |
| Both set | `--rps` takes priority |

### Important

- Rate limiting applies to **all** HTTP requests (page fetches and broken link checks)
- Context cancellation immediately stops waiting, no hang on shutdown
- The interval is measured from the start of each request

## Retries

The crawler automatically retries failed requests for temporary errors.

### Usage

```bash
bin/hexlet-go-crawler --retries=3 https://example.com
```

### Retryable Conditions

Retries are performed only for **temporary errors**:

| Type | Codes/Errors |
|------|--------------|
| HTTP status | 429 (Too Many Requests), 500, 502, 503, 504 |
| Network | Connection timeouts, temporary network failures |

**Non-retryable** errors (e.g., 404 Not Found, 403 Forbidden) are reported immediately.

### Retry Logic

- **Total attempts** = `retries + 1` (initial request + retry attempts)
- **Delay between retries**: 100ms (prevents request bursts)
- **Report result**: reflects the **last** attempt (success or failure)

### Examples

| Scenario | Retries | Attempts | Result |
|----------|---------|----------|--------|
| Success on first try | 2 | 1 | Success |
| 503 → 503 → 200 | 2 | 3 | Success |
| 503 → 503 → 503 | 2 | 3 | Error (503) |
| 404 | 2 | 1 | Error (404, no retry) |

### Context Cancellation

If the context is canceled during retry wait, the crawler stops immediately without hanging.

## Assets

The crawler collects information about static assets (images, scripts, stylesheets) on each page.

### Asset Structure

Each asset in the report contains:

```json
{
  "url": "https://example.com/static/logo.png",
  "type": "image",
  "status_code": 200,
  "size_bytes": 12345,
  "error": ""
}
```

### Supported Types

| Type | Elements |
|------|----------|
| `image` | `<img src="...">` |
| `script` | `<script src="...">` |
| `style` | `<link rel="stylesheet" href="...">` |

### Size Detection

- If the server provides `Content-Length` header, it is used directly
- Otherwise, size is calculated from the response body
- If size cannot be determined, `size_bytes` is `0`

### Caching

Assets are cached by URL. If the same asset appears on multiple pages, it is fetched only once. This reduces network requests and ensures consistent data across pages.

### Error Handling

- HTTP errors (4xx, 5xx) are reported with `status_code` and `error` message
- Network errors are reported in the `error` field
- All fields are always present, even on errors