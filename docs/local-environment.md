# Local environment for running & testing scanners

This guide lets you run the **k8s-assets** and **keppel** scanners end-to-end on your
laptop — no remote cluster or registry required. Both scanners report into a local
heureka backend running in docker-compose.

- **k8s-assets** scans a Kubernetes cluster. Locally we use a [kind](https://kind.sigs.k8s.io/)
  cluster seeded with sample workloads from the existing `scanner/playground` Helm chart.
- **keppel** is *not* a Kubernetes scanner — it is an HTTP client for a Keppel image
  registry fronted by OpenStack Keystone. Locally we point it at an in-repo **mock**
  (`scanner/keppel/mock`) that emulates Keystone v3 auth and the Keppel endpoints.

## Prerequisites

- Docker, [`kind`](https://kind.sigs.k8s.io/), [`helm`](https://helm.sh/), `go` (version per
  the scanner modules), and a free set of ports: `:80` (heureka), `:3306` (MariaDB),
  `:6379` (valkey), `:8080`/`:8443` (keppel mock).

## 1. Start the backend + cluster

```bash
make local-up
```

This runs `compose-prepare` + `compose-up` (heureka + MariaDB + valkey, GraphQL at
`http://localhost/query`, playground at `http://localhost/playground`), then `kind-up`
(creates the `heureka-local` cluster and exports its kubeconfig to
`local/kind/kubeconfig`) and `playground-up` (installs the sample namespaces/pods).

> There is **no auth locally** — the compose backend sets no `AUTH_TOKEN_SECRET`, and the
> scanners never send an API token. `HEUREKA_API_TOKEN` is a no-op here.

## 2. Run the k8s-assets scanner

```bash
make scan-k8s-assets-local
```

Runs the scanner on the host (`KUBE_CONFIG_TYPE=default`) against the kind kubeconfig and
reports to local heureka. Image pulls into kind are **not** required — the scan reads pod
specs even if pods are in `ImagePullBackOff`.

Verify in the GraphQL playground (`http://localhost/playground`): a completed
`scannerRuns` entry tagged `k8s-assets-local-kind`, and `services` / `componentInstances`
/ `components` populated from the playground pods (nginx, mariadb, mongo, redis, nats, httpd).

## 3. Run the keppel scanner

The keppel scanner only looks at **components that already exist in heureka**, so run
k8s-assets first (step 2) to create image-shaped components.

In one terminal, start the mock:

```bash
make keppel-mock-up      # Keystone :8080 (http), Keppel :8443 (https, self-signed)
```

In another terminal, run the scanner against it:

```bash
make scan-keppel-local
```

Verify: a completed `scannerRuns` entry tagged `Keppel`, and `componentVersions` /
`issueMatches` carrying the mock's CVEs (e.g. `CVE-2021-44228`).

## 4. Tear down

```bash
make local-down          # preserves the local DB
```

Use `local-down` (or `compose-down`) — **not** `make stop` or `make clean`, which delete
`./.db/data` / compose volumes and wipe your scanned data.

## How the keppel mock stays in sync with the real API

The mock is a Go program **inside the keppel module** (`scanner/keppel/mock`), so it imports
the scanner's own response structs from `scanner/keppel/models`. Three layers keep it honest:

1. **Compile-time coupling.** Responses are built from `models` structs and the upstream
   trivy types — a breaking change to those types is a compile error in the mock.
2. **CI drift guard.** `fixtures_test.go` asserts the built-in responses round-trip and that
   any committed override fixture still unmarshals into its model type. Run with
   `cd scanner/keppel && go test ./mock/...`.
3. **Record mode.** When the real API changes and you have staging credentials, refresh the
   fixtures from reality:

   ```bash
   cd scanner/keppel/mock
   KEPPEL_RECORD=1 \
     RECORD_IDENTITY_UPSTREAM=https://identity.staging.example.com \
     RECORD_KEPPEL_UPSTREAM=https://keppel.staging.example.com \
     KEPPEL_USERNAME=... KEPPEL_PASSWORD=... KEPPEL_USER_DOMAIN=... \
     KEPPEL_DOMAIN=... KEPPEL_PROJECT=... \
     go run .   # (or: make keppel-mock-record with the same env)
   ```

   Then point the keppel scanner at the mock and run it once (step 3). Each Keppel response
   is written to `scanner/keppel/mock/testdata/fixtures/<name>.json`; commit the updated
   `models.go` (if any) together with the refreshed fixtures.

   Update workflow: edit `models.go` if the shape changed → `go test ./mock/...` (fails on
   drift) → `make keppel-mock-record` (or hand-edit fixtures) → re-run the test → commit.

When no override fixture is present, the mock serves a built-in default; an override file
takes precedence. Manifest response headers (`X-Keppel-Vulnerability-Status`,
`Docker-Content-Digest`) are set statically in serve mode.

## Notes & gotchas

- **TLS asymmetry**: Keystone (`IDENTITY_ENDPOINT`) is plain HTTP because gophercloud
  verifies TLS there; Keppel (`https://$KEPPEL_FQDN`) is HTTPS self-signed because the
  scanner skips verification only for its own Keppel client.
- The keppel Helm chart's CronJob has a placeholder `args` (`date; echo Hello…`) that
  overrides the real command, so you can't fall back to in-cluster Jobs without fixing it —
  host `go run` (the flow above) avoids this.
- `KUBE_CONFIG_CONTEXT` is ignored with `KUBE_CONFIG_TYPE=default`; the kind kubeconfig's
  current-context (set by `kind export kubeconfig`) is what matters.
- **Rebuild the backend after pulling GraphQL schema changes**: the compose `heureka-app`
  image bakes in the generated schema, so after a rebase that adds fields (e.g.
  `ComponentVersion.latestVersion`) run `make compose-build && make compose-restart` or the
  scanner's mutations will fail against the stale running server. (Restart re-seeds the DB
  when `SEED_MODE=true`, so re-run the scanners afterwards.)
- **`scan-keppel-local` sets `KEPPEL_FQDN_ALIASES=docker.io`** so the scanner's `OwnsRegistry`
  check passes for the `docker.io/*` components created by k8s-assets — that's what exercises
  the available-version-update feature (`ListTags` → highest semver). The mock's `_manifests`
  listing returns tags `1.0.0 / 2.0.5 / 2.3.1 / latest`, so `latestVersion` becomes `2.3.1`.

