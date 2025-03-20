# ComponentVersionIssue

The `component_version_issue.json` mapping is utilized to validate the ordering of `ComponentVersion` instances based on their Severity (Rating). The descending order prioritizes as follows:
- Highest count of Critical issues
- Highest count of High issues
- Highest count of Medium issues
- Highest count of Low issues
- Highest count of None issues

## Mapping

| ComponentVersionId | Critical | High | Medium | Low | None |
|--------------------|----------|------|--------|-----|------|
| 1                  | 0        | 2    | 2      | 2   | 1    |
| 2                  | 2        | 0    | 2      | 0   | 1    |
| 3                  | 2        | 2    | 2      | 2   | 1    |
| 4                  | 0        | 0    | 0      | 2   | 1    |
| 5                  | 0        | 0    | 2      | 2   | 1    |
| 6                  | 0        | 1    | 2      | 2   | 1    |
| 7                  | 1        | 1    | 0      | 0   | 0    |
| 8                  | 2        | 1    | 0      | 0   | 0    |
| 9                  | 0        | 0    | 0      | 0   | 1    |
| 10                 | 0        | 0    | 0      | 1   | 1    |

### DESC Ordered

| ComponentVersionId | Critical | High | Medium | Low | None |
|--------------------|----------|------|--------|-----|------|
| 3                  | 2        | 2    | 2      | 2   | 1    |
| 8                  | 2        | 1    | 0      | 0   | 0    |
| 2                  | 2        | 0    | 2      | 0   | 1    |
| 7                  | 1        | 1    | 0      | 0   | 0    |
| 1                  | 0        | 2    | 2      | 2   | 1    |
| 6                  | 0        | 1    | 2      | 2   | 1    |
| 5                  | 0        | 0    | 2      | 2   | 1    |
| 4                  | 0        | 0    | 0      | 2   | 1    |
| 10                 | 0        | 0    | 0      | 1   | 1    |
| 9                  | 0        | 0    | 0      | 0   | 1    |

### ASC Ordered

| ComponentVersionId | Critical | High | Medium | Low | None |
|--------------------|----------|------|--------|-----|------|
| 9                  | 0        | 0    | 0      | 0   | 1    |
| 10                 | 0        | 0    | 0      | 1   | 1    |
| 4                  | 0        | 0    | 0      | 2   | 1    |
| 5                  | 0        | 0    | 2      | 2   | 1    |
| 6                  | 0        | 1    | 2      | 2   | 1    |
| 1                  | 0        | 2    | 2      | 2   | 1    |
| 7                  | 1        | 1    | 0      | 0   | 0    |
| 2                  | 2        | 0    | 2      | 0   | 1    |
| 8                  | 2        | 1    | 0      | 0   | 0    |
| 3                  | 2        | 2    | 2      | 2   | 1    |