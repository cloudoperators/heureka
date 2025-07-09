# Cache Function Documentation

Heureka backend project supports a pluggable caching mechanism to improve performance, especially when accessing the MariaDB database. The caching is handled via a unified interface and can be enabled or disabled using environment variables.

## Cache Types

There are **three types of cache** supported:

1. **NoCache**
2. **InMemoryCache**
3. **RedisCache**

### Cache Selection Logic

The type of cache used is determined by the following environment variables:

- `CACHE_TTL_MSEC`: TTL (time to live) for a cache entry, in milliseconds.
- `CACHE_REDIS_URL`: Redis host in the form `domain:port`.

#### Selection Rules:

| Condition | Cache Type |
|----------|------------|
| `CACHE_TTL_MSEC=0` | NoCache |
| `CACHE_TTL_MSEC > 0` AND `CACHE_REDIS_URL` is set | RedisCache |
| `CACHE_TTL_MSEC > 0` AND `CACHE_REDIS_URL` is **not** set | InMemoryCache |

---

## CallCached Function

`CallCached()` is a generic cache wrapper function that adds caching capability to any function returning a single serializable object and an error.

### Function Signature

```go
CallCached(cache Cache, name string, fn interface{}, args ...interface{}) (interface{}, error)
```

### Parameters

1. `cache`: A cache implementation (e.g., Redis, in-memory, or no-op).
2. `name`: A string identifier for the wrapped function (not inferred at runtime).
3. `fn`: The actual function to be wrapped.
4. `args...`: Arguments to pass to the wrapped function.

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
