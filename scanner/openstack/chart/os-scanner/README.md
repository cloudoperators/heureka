OS Scanner for Heureka
==============================

This scanner is used to perform scans for the heureka project.

## Usage
```bash
$ os-manual --help
```

## Helm Chart
Usage:
```bash 
helm upgrade --install --namespace heureka os-manual heureka/scanner/openstack/chart/os-scanner/
```

#### Values
In the `values.yaml` file, you can configure the following values:
- `scanner.api_token`: The token used to authenticate the scanner.
- `scanner.heureka_url`: The URL of the Heureka API.
- `scanner.config_mount_path`: The path of the scanner config file inside the pod (e.g. "/etc/heureka/scanner/OS/config")
- `scanner.schedule`: The cronjob schedule string (e.g. "0 * * * *") that defines when the scanner should run.
- `scanner.auth_url` : OS auth url
- `scanner.domain_name` : OS domain name
- `scanner.password` : OS password
- `scanner.project_id` : OS project id
- `scanner.project_name` : OS project name
- `scanner.region_name` : OS region name
- `scanner.username` : OS username
- `scanner.config_mount_path` : OS mount path