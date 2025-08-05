# 📘 CPU Profiling in Heureka Go Application (Docker)

This guide explains how to enable, run, and analyze CPU profiling for the Heureka Go application running in a Docker container.

---

## ⚙️ Enable CPU Profiling

1. **Set the `CPU_PROFILER_FILE_PATH` environment variable** in your `docker-compose.yml` or container spec:

```yaml
environment:
  - CPU_PROFILER_FILE_PATH=/home/nonroot/cpu.prof
```

> When this environment variable is set (i.e., not empty), the application will start profiling CPU usage and write the data to the specified file.

---

## 🛑 Stop Profiling & Finalize the File

To safely stop profiling and finalize the profile output:

```bash
docker stop heureka-app
```

> This ensures that profiling is stopped cleanly and the `cpu.prof` file is properly flushed and saved.

---

## 📤 Copy Profile File from Container

After the container stops, copy the profile data file from the container to your local system:

```bash
docker cp heureka-app:/home/nonroot/cpu.prof /tmp/cpu.prof
```

> Replace `/home/nonroot/cpu.prof` with your configured path if different.

---

## 🔍 Analyze the Profile

Use `go tool pprof` to inspect the profile:

```bash
go tool pprof --focus=internal/ --hide=runtime build/heureka /tmp/cpu.prof
```

This will launch an interactive terminal interface where you can:

- Type `top` to view top consumers
- Type `web` to generate a call graph (requires Graphviz)
- Filter and inspect specific functions or packages

---

## ✅ Notes

- The profile path **must be writable by the app user** (`/home/nonroot` is a good default).
- The profiling will only activate if `CPU_PROFILER_FILE_PATH` is **set and non-empty**.
- Ensure your Go binary (`build/heureka`) includes debug symbols (not stripped).
