Keppel Scanner for Heureka
==============================

This scanner is used to perform scans for the heureka project.

## Usage
```bash
$ keppel-scanner --help
```

## Helm Chart
Usage:
```bash 
helm upgrade --install --namespace heureka keppel-scanner heureka/scanner/keppel/chart/keppel-scanner/
```

#### Values
In the `values.yaml` file, you can configure the following values:
- `scanner.api_token`: The token used to authenticate the scanner.
- `scanner.heureka_url`: The URL of the Heureka API.
- `scanner.config_mount_path`: The path of the scanner config file inside the pod (e.g. "/etc/heureka/scanner/keppel/config")
- `scanner.schedule`: The cronjob schedule string (e.g. "0 * * * *") that defines when the scanner should run.