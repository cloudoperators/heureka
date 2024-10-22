# Kubernetes Assets Scanner

The Kubernetes Assets Scanner is a tool designed to scan and collect information about services, pods, and containers running in a Kubernetes cluster. It processes the collected data and reports findings to a GraphQL API (presumably Heureka).

## Prerequisites

- Go 1.15 or later
- Access to a Kubernetes cluster
- Heureka system for reporting findings

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/cloudoperators/heureka.git
   cd scanners/k8s-assets
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Configuration

The scanner is configured using environment variables. Set the following variables before running the scanner:

- `HEUREKA_LOG_LEVEL`: Set the log level (default: "debug")
- `HEUREKA_KUBE_CONFIG_PATH`: Path to kubeconfig file (default: "~/.kube/config")
- `HEUREKA_KUBE_CONFIG_CONTEXT`: Kubernetes context to use
- `HEUREKA_KUBE_CONFIG_TYPE`: Type of Kubernetes config (default: "oidc")
- `HEUREKA_SUPPORT_GROUP_LABEL`: Label for support group (default: "ccloud/support-group")
- `HEUREKA_SERVICE_NAME_LABEL`: Label for service name (default: "ccloud/service")
- `HEUREKA_SCANNER_TIMEOUT`: Timeout for the scanner (default: "30m")
- `HEUREKA_HEUREKA_URL`: URL of the Heureka system for reporting findings
- `HEUREKA_CLUSTER_NAME`: Name of the cluster being scanned
- `HEUREKA_CLUSTER_REGION`: Region of the cluster being scanned

Example:

```bash
export HEUREKA_LOG_LEVEL=debug
export HEUREKA_KUBE_CONFIG_PATH=~/.kube/config
export HEUREKA_KUBE_CONFIG_CONTEXT=my-cluster-context
export HEUREKA_KUBE_CONFIG_TYPE=oidc
export HEUREKA_SUPPORT_GROUP_LABEL=ccloud/support-group
export HEUREKA_SERVICE_NAME_LABEL=ccloud/service
export HEUREKA_SCANNER_TIMEOUT=30m
export HEUREKA_HEUREKA_URL=https://heureka.example.com
export HEUREKA_CLUSTER_NAME=my-cluster
export HEUREKA_CLUSTER_REGION=us-west-1
```

## Usage

To run the Kubernetes Assets Scanner:

```bash
go run main.go
```

The scanner will start processing namespaces, services, pods, and containers, and report findings to the configured Heureka system.
