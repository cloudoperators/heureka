# NVD Scanner

## Description

The NVD Scanner fetches CVE data from the [NIST National Vulnerability Database (NVD) API 2.0](https://nvd.nist.gov/developers/vulnerabilities) and writes it into Heureka as `Issue` records of type `Vulnerability`.

For each CVE the scanner stores:

- **Description** — the English-language description from NVD
- **CVSS severity vector** — used to populate the linked `IssueVariant`
- **CISA KEV signal** — if the CVE is in [CISA's Known Exploited Vulnerabilities (KEV) catalog](https://www.cisa.gov/known-exploited-vulnerabilities-catalog), three additional fields are set on the `Issue`:
  - `knownExploited: true`
  - `knownExploitedAddedDate` — date the CVE was added to the KEV catalog (`cisaExploitAdd` from NVD)
  - `knownExploitedDueDate` — CISA's remediation deadline for federal agencies (`cisaActionDue` from NVD)

## Run Modes

The scanner supports two mutually-exclusive modes selected by environment variables:

### Initial ingest (publication-date window)

Set `HEUREKA_NVD_START_DATE` to ingest CVEs published in a specific time range. The scanner pages through the window in 2-month batches automatically.

```bash
HEUREKA_NVD_START_DATE=2024-01-01 \
HEUREKA_NVD_END_DATE=2024-12-31 \
go run .
```

Each CVE is created fresh (`createIssue` + `createIssueVariant`). If a CVE already exists the mutation returns a conflict error and the CVE is skipped — use update mode for re-ingestion.

### Update mode (last-modified window)

Set `HEUREKA_NVD_UPDATE_MODE=true` to check for CVEs modified in the last N days (default: 7). This is the correct mode for scheduled production runs and for picking up newly KEV-listed CVEs.

```bash
HEUREKA_NVD_UPDATE_MODE=true \
HEUREKA_NVD_REVIEW_INTERVAL_DAYS=7 \
go run .
```

For each CVE in the window the scanner:
1. Looks up the existing `Issue` in Heureka by CVE ID.
2. If not found — creates it (same as initial ingest).
3. If found — compares description, CVSS vector, and `knownExploited` flag. Writes an update only when something changed.

> When CISA adds a CVE to the KEV catalog, NVD bumps its `lastModified` timestamp. This causes the CVE to appear in the next update-mode window, and `knownExploited` is flipped from `false` to `true` automatically.

## Prerequisites

- Go 1.22 or later
- A running Heureka instance
- An [NVD API key](https://nvd.nist.gov/developers/request-an-api-key) (required; unauthenticated requests are heavily rate-limited)

## Configuration

All environment variables are prefixed with `HEUREKA_`.

### Scanner (`HEUREKA_NVD_*`)

| Variable | Required | Default | Description |
|---|---|---|---|
| `HEUREKA_NVD_API_URL` | yes | — | NVD CVE API base URL (`https://services.nvd.nist.gov/rest/json/cves/2.0`) |
| `HEUREKA_NVD_API_KEY` | yes | — | NVD API key |
| `HEUREKA_NVD_START_DATE` | no | `""` | Start of publication-date window (`YYYY-MM-DD`). Activates initial-ingest mode. |
| `HEUREKA_NVD_END_DATE` | no | `""` | End of publication-date window (`YYYY-MM-DD`). |
| `HEUREKA_NVD_UPDATE_MODE` | no | `false` | Enable update mode (last-modified window). |
| `HEUREKA_NVD_REVIEW_INTERVAL_DAYS` | no | `7` | Days back from today for the update-mode window. |
| `HEUREKA_NVD_RESULTS_PER_PAGE` | no | `2000` | NVD API page size (max 2000). |
| `HEUREKA_NVD_RATE_LIMIT` | no | `1.666` | Max NVD requests per second. |
| `HEUREKA_NVD_RATE_BURST` | no | `50` | Burst allowance for the NVD rate limiter. |

### Processor (`HEUREKA_HEUREKA_*`)

| Variable | Required | Default | Description |
|---|---|---|---|
| `HEUREKA_HEUREKA_URL` | yes | — | Heureka GraphQL endpoint (e.g. `http://localhost/query`) |
| `HEUREKA_HEUREKA_RATE_LIMIT` | no | `10000` | Max Heureka requests per second (client-side). Keep below the server's `GQL_HTTP_RATE_LIMIT` (default 100). |
| `HEUREKA_HEUREKA_RATE_BURST` | no | `10000` | Burst allowance for the Heureka rate limiter. |
| `HEUREKA_ISSUE_REPOSITORY_NAME` | no | `nvd` | Name of the `IssueRepository` record created in Heureka. |
| `HEUREKA_ISSUE_REPOSITORY_URL` | no | `https://nvd.nist.gov/` | URL of the issue repository. |
| `HEUREKA_CVE_DETAILS_URL` | no | `https://nvd.nist.gov/vuln/detail/` | Base URL prepended to CVE IDs for `IssueVariant.externalUrl`. |

## Usage

```bash
# Initial ingest for a specific year
HEUREKA_NVD_API_URL=https://services.nvd.nist.gov/rest/json/cves/2.0 \
HEUREKA_NVD_API_KEY=<your-key> \
HEUREKA_HEUREKA_URL=http://localhost/query \
HEUREKA_NVD_START_DATE=2024-01-01 \
HEUREKA_NVD_END_DATE=2024-12-31 \
go run .

# Scheduled update run (e.g. daily cron)
HEUREKA_NVD_API_URL=https://services.nvd.nist.gov/rest/json/cves/2.0 \
HEUREKA_NVD_API_KEY=<your-key> \
HEUREKA_HEUREKA_URL=http://heureka.example.com/query \
HEUREKA_NVD_UPDATE_MODE=true \
go run .
```

## Testing

Run the unit and integration tests from the repository root:

```bash
go test ./scanner/nvd/...
```

## Helm Deployment

See [chart/nvd-scanner/README.md](chart/nvd-scanner/README.md) for Helm chart usage and values.
