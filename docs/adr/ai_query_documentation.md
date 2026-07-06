# AI Natural Language Query

Heureka exposes a natural language query endpoint (`POST /ai/query`) that translates a plain English question into a GraphQL query, validates it, executes it, and returns the results. The endpoint is backed by a provider-agnostic LLM abstraction so the underlying model can be swapped without changing application code.

---

## Enabling the Endpoint

The AI endpoint is **disabled by default**. Set `AI_ENABLE=true` to register the route at startup.

When `AI_ENABLE=false` (or unset), the `/ai/query` route is never registered and no LLM client is constructed.

---

## Configuration

| Environment Variable   | Default       | Description |
|------------------------|---------------|-------------|
| `AI_ENABLE`            | `false`       | Set to `true` to enable the `/ai/query` endpoint. |
| `AI_PROVIDER`          | `anthropic`   | LLM API format: `anthropic` or `openai`. |
| `AI_SAP_PROXY_URL`     | _(empty)_     | Base URL of the LLM API (e.g. `https://host/anthropic/v1`). |
| `AI_SAP_PROXY_MODEL`   | _(empty)_     | Model identifier to pass to the LLM API. |
| `AI_SAP_PROXY_TOKEN`   | _(empty)_     | Bearer token for authenticating with the LLM API. |
| `AI_HTTP_RATE_LIMIT`   | `5.0`         | Token-bucket refill rate — maximum sustained requests per second per IP. |
| `AI_HTTP_RATE_BURST`   | `5`           | Maximum burst size — peak requests allowed before throttling kicks in. |

### Provider selection

`AI_PROVIDER=anthropic` (default) expects an Anthropic Messages API format (`POST /messages`, top-level `system` field, `content[]` response).

`AI_PROVIDER=openai` expects an OpenAI-compatible Chat Completions format (`POST /chat/completions`, `messages[]` with roles, `choices[]` response).

---

## API

### `POST /ai/query`

**Request body**

```json
{
  "question": "Show me all critical severity vulnerabilities with their names and descriptions"
}
```

**Success response** `200 OK`

```json
{
  "generatedQuery": "query { Vulnerabilities(filter: { severity: [Critical] }) { edges { node { name description } } } }",
  "data": { ... }
}
```

**Error response** (all error cases)

```json
{
  "errors": [{ "message": "..." }]
}
```

The `errors` array mirrors the GraphQL error shape so clients can handle both the `/query` and `/ai/query` endpoints uniformly.

| HTTP Status | Condition |
|-------------|-----------|
| `400`       | Malformed request body, prompt injection detected, or generated query failed validation. |
| `422`       | The question cannot be answered with a read-only query (LLM returned UNSUPPORTED). |
| `429`       | Rate limit exceeded for the calling IP. |
| `500`       | LLM call failed or internal GraphQL execution error. |

---

## Security

The endpoint has three independent layers of protection against abuse and prompt injection.

### 1. Regex pre-screen

Before the LLM is called, the user question is matched against a set of known injection phrases:

- `ignore previous/prior/above instructions`
- `new instructions`
- `you are now`
- `pretend you are / pretend to be`
- `act as`
- `forget everything / forget all / forget your`
- `disregard`
- `override your rules / override your instructions`
- `system prompt`

A match returns `400` immediately without consuming any LLM tokens.

### 2. System prompt hardening

The LLM is instructed via a hardened system prompt that:

- Limits output to `query` operations only.
- Restricts fields and types to those present in the embedded schema.
- Instructs the model to respond with `UNSUPPORTED` for any question it cannot answer as a read-only query.
- Instructs the model to ignore any in-message attempts to override these rules.

The user question is always passed as a separate `user` message, never interpolated into the system prompt, keeping the trust boundary intact.

### 3. AST validation (hard enforcement)

After the LLM response is received the generated query is validated in Go code regardless of what the model produced:

- Keyword check — rejects any query containing `mutation` or `subscription` (case-insensitive).
- Introspection block — rejects queries containing `__schema` or `__type`.
- Schema parse — the query is parsed against the embedded `.graphqls` schema using `gqlparser`; invalid queries are rejected.
- Operation type check — the parsed AST is inspected; only `query` operations are allowed.

This means even a fully jailbroken model cannot cause a mutation to be executed — the Go validation layer is independent of the LLM.

---

## Rate Limiting

The `/ai` route group has its own dedicated IP-based rate limiter, separate from and tighter than the global GraphQL rate limiter.

| Limiter         | Default rate | Default burst |
|-----------------|-------------|---------------|
| GraphQL (`/query`) | 100 req/s   | 100           |
| AI (`/ai/query`)   | 5 req/s     | 5             |

When the limit is exceeded the endpoint returns `429 Too Many Requests` with `Retry-After` and `X-RateLimit-*` headers.

---

## LLM Provider Abstraction

All LLM calls go through the `llm.Client` interface:

```go
type Client interface {
    Complete(ctx context.Context, systemPrompt, userMessage string) (string, error)
}
```

Two implementations are bundled:

| Implementation   | Selected when        | API format |
|------------------|----------------------|------------|
| `AnthropicClient`| `AI_PROVIDER=anthropic` (default) | Anthropic Messages API |
| `SAPProxyClient` | `AI_PROVIDER=openai` | OpenAI Chat Completions |

Adding a new provider means implementing the `Client` interface and adding a case to the provider switch in `internal/server/server.go`.

---

## Schema Context

The full GraphQL schema is embedded at build time using `//go:embed` in `internal/api/graphql/schema.go` and injected verbatim into the LLM system prompt. This means the model always works against the current schema without any manual synchronisation step.
