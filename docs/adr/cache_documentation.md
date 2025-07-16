# Cache Function Documentation

Heureka backend project supports a pluggable caching mechanism to improve performance, especially when accessing the MariaDB database. The caching is handled via a unified interface and can be enabled or disabled using environment variables.

## Cache Types

There are **three types of cache** supported:

1. **NoCache**
2. **InMemoryCache**
3. **ValkeyCache**

### Cache Selection Logic

The type of cache used is determined by the following environment variables:

- `CACHE_ENABLE`: bool value to enable caching
- `CACHE_VALKEY_URL`: Valkey host in the form `domain:port`.

#### Selection Rules:

| Condition | Cache Type |
|----------|------------|
| `CACHE_ENABLE=false` | NoCache |
| `CACHE_ENABLE=true` AND `CACHE_VALKEY_URL` is set | ValkeyCache |
| `CACHE_ENABLE=true` AND `CACHE_VALKEY_URL` is **not** set | InMemoryCache |

---

## CallCached Function

`CallCached()` is a generic cache wrapper function that adds caching capability to any function returning a single serializable object and an error.

### Function Signature

```go
CallCached(cache Cache, ttl time.Duration, name string, fn interface{}, args ...interface{}) (interface{}, error)
```

### Parameters

1. `cache`: A cache implementation (e.g., Valkey, in-memory, or no-op).
2. `ttl`: Caching time to live.
3. `name`: A string identifier for the wrapped function (not inferred at runtime).
4. `fn`: The actual function to be wrapped.
5. `args...`: Arguments to pass to the wrapped function.

### Notes

- **Function Name (Param 2)**: To avoid reflection or runtime overhead, function names are explicitly passed as strings.
- A static check is enforced via a `go generate` step (`make check`) to verify that all `CallCached` usages are correctly formed.

---

## Cache Behavior

- **Cache Key**: A unique string based on the function name and parameters. It can be either no enoded json object containing name of the function and serialized call function parameters or mentioned json object encoded using one of: HEX, Base64, SHA256, SHA512 (SHA encoding is non-reversible).
- **Cache Value**: A JSON-serialized string representation of the returned object.
- **TTL**: Used to expire old entries. Expired entries are removed/replaced upon access.

### Background Refresh

On cache **hit**, the cached value is returned immediately. In the **background**, a fresh value is fetched from the source (e.g., database), and the cache is updated. This ensures faster responses while keeping the cache up-to-date.

---

## Statistics

Each cache implementation provides the following statistics:

```go
type Cache interface {
    GetStat() Stat
}
```

Where `Stat` contains:

- `Hit`: Number of cache hits.
- `Miss`: Number of cache misses.

---

## Thread Safety

The cache implementations are **thread-safe**.

---

## Context Support

Context support is **not implemented**, as the handlers using the cache currently do not propagate `context.Context`.

---

## Example Use Case

The `CallCached()` function is primarily used to wrap **MariaDB data access functions**, which are frequently queried with the same parameters. Caching shortens response times significantly.

---

## Cache Monitor

A **cache monitor** is available to track and log cache statistics periodically. It is controlled via the environment variable `CACHE_MONITOR_MSEC`.

### Configuration

- If `CACHE_MONITOR_MSEC` is **not set** or set to a value **≤ 0**, the cache monitor is **disabled**.
- If `CACHE_MONITOR_MSEC` is set to a **positive integer**, cache monitoring is **enabled**.
- The value defines the **interval in milliseconds** at which the monitor logs current cache statistics (hits and misses).

This feature is useful for tracking cache effectiveness over time in production environments.

# 🔄 Cache Background Refresh Throttling

The cache system supports background refresh of cached entries after every cache hit. To avoid overloading the database, two independent throttling mechanisms can be configured using environment variables:

1. **Concurrency-based throttling**
2. **Rate-based throttling**

Both can be enabled individually or together depending on your requirements.

---

## 🧵 1. Concurrency-Based Throttling

Controlled via the `CACHE_MAX_DB_CONCURRENT_REFRESHES` environment variable, this mechanism limits how many background refresh goroutines can run at the same time.

### Environment Variable: `CACHE_MAX_DB_CONCURRENT_REFRESHES`

| Value  | Behavior                                                                 |
|--------|--------------------------------------------------------------------------|
| `-1`   | **Unlimited** – background refresh is **enabled for every hit** with **no limit** on concurrency. |
| `0`    | **Disabled** – background refresh is **completely turned off**.          |
| `> 0`  | **Throttled** – background refresh is enabled with a limit of **N** concurrent refreshes. Any refreshes above this limit are **discarded** silently. |

### Examples

- `CACHE_MAX_DB_CONCURRENT_REFRESHES=-1`: Refresh data in background after every cache hit without any concurrency limit.
- `CACHE_MAX_DB_CONCURRENT_REFRESHES=0`: Completely disables background refresh.
- `CACHE_MAX_DB_CONCURRENT_REFRESHES=5`: Allows up to 5 background refresh jobs at the same time. Extra jobs are dropped.

### Use Cases

- You want fresh data but need to protect the DB from spikes of load.
- You want to temporarily disable background jobs.
- You need to guarantee a maximum number of simultaneous DB refreshes.

---

## ⏱ 2. Rate-Based Throttling

Controlled via the `CACHE_THROTTLE_INTERVAL_MSEC` and `CACHE_THROTTLE_PER_INTERVAL` environment variables, this mechanism limits the frequency of background refreshes over time.

### Environment Variable: `CACHE_THROTTLE_INTERVAL_MSEC`

- **Type**: `int`
- **Description**: Defines the time interval (in milliseconds) in which background DB refresh requests are allowed.
- If this is not set or is set to `0`, **rate-based throttling is disabled** (default behavior).

### Environment Variable: `CACHE_THROTTLE_PER_INTERVAL`

- **Type**: `int`
- **Default**: `1`
- **Description**: Defines how many background DB refresh requests are allowed per interval.

### Behavior Summary

- If `CACHE_THROTTLE_INTERVAL_MSEC` is **not set or is 0**, rate-based throttling is **disabled**.
- If it is set to a positive number:
  - Only `CACHE_THROTTLE_PER_INTERVAL` background refreshes are allowed per `CACHE_THROTTLE_INTERVAL_MSEC` window.
  - Additional refresh attempts during that interval are **skipped/discarded**.

### Example Scenarios

#### Throttling Disabled

```env
CACHE_THROTTLE_INTERVAL_MSEC=0
```

→ No throttling is applied. Every cache hit may trigger a background refresh.

#### One Refresh Per Second

```env
CACHE_THROTTLE_INTERVAL_MSEC=1000
CACHE_THROTTLE_PER_INTERVAL=1
```

→ At most one background refresh will be allowed every second.

#### Two Refreshes Every 500ms

```env
CACHE_THROTTLE_INTERVAL_MSEC=500
CACHE_THROTTLE_PER_INTERVAL=2
```

→ At most two refreshes are allowed every 500 milliseconds.

---

## 🔧 Combined Usage

You can combine both mechanisms for finer control. For example:

```env
CACHE_MAX_DB_CONCURRENT_REFRESHES=3
CACHE_THROTTLE_INTERVAL_MSEC=1000
CACHE_THROTTLE_PER_INTERVAL=2
```

→ At most 2 refreshes per second will be attempted, and no more than 3 will run at the same time.

---

**Note:** Skipped or discarded refreshes (due to either limit) are silently ignored to preserve performance and avoid queueing or retry storms.
