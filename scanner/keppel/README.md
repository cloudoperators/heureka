# Keppel Scanner

The Keppel scanner is a tool designed to scan container images stored in a [Keppel](https://github.com/sapcc/keppel) registry. It retrieves information about accounts, repositories, and manifests, and processes vulnerability reports for each image.

## Prerequisites

- Go 1.16 or later
- Access to a Keppel registry
- Heureka system for reporting findings

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/cloudoperators/heureka.git
   cd scanner/keppel
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Configuration

The scanner is configured using environment variables. Set the following variables before running the scanner:

- `HEUREKA_LOG_LEVEL`: Set the log level (default: "debug")
- `HEUREKA_KEPPEL_FQDN`: Fully Qualified Domain Name of the Keppel registry
- `HEUREKA_KEPPEL_USERNAME`: Username for Keppel authentication
- `HEUREKA_KEPPEL_PASSWORD`: Password for Keppel authentication
- `HEUREKA_KEPPEL_DOMAIN`: Domain for Keppel authentication
- `HEUREKA_KEPPEL_PROJECT`: Project for Keppel authentication
- `HEUREKA_IDENTITY_ENDPOINT`: Identity endpoint for authentication
- `HEUREKA_HEUREKA_URL`: URL of the Heureka system for reporting findings

Example:

```bash
export HEUREKA_LOG_LEVEL=debug
export HEUREKA_KEPPEL_FQDN=keppel.example.com
export HEUREKA_KEPPEL_USERNAME=myusername
export HEUREKA_KEPPEL_PASSWORD=mypassword
export HEUREKA_KEPPEL_DOMAIN=mydomain
export HEUREKA_KEPPEL_PROJECT=myproject
export HEUREKA_IDENTITY_ENDPOINT=https://identity.example.com
export HEUREKA_HEUREKA_URL=https://heureka.example.com
```

## Usage

To run the Keppel scanner:

```bash
go run main.go
```

The scanner will start processing accounts, repositories, and manifests, and report findings to the configured Heureka system.

