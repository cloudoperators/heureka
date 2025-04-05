# Service

The `component_instance.json` mapping is utilized to validate the ordering of `Services` instances based on their Severity (Rating). The descending order prioritizes as follows:
- Highest count of Critical issues
- Highest count of High issues
- Highest count of Medium issues
- Highest count of Low issues
- Highest count of None issues

## Mapping

| Service | Critical | High | Medium | Low | None |
|---------|----------|------|--------|-----|------|
|1        | 1        | 0    | 0      | 1   | 0    |
|2        | 0        | 0    | 1      | 1   | 0    |
|3        | 1        | 0    | 0      | 0   | 1    |
|4        | 0        | 1    | 1      | 0   | 0    |
|5        | 0        | 1    | 0      | 0   | 1    |


### DESC Ordered

| Service | Critical | High | Medium | Low | None |
|---------|----------|------|--------|-----|------|
|1        | 1        | 0    | 0      | 1   | 0    |
|3        | 1        | 0    | 0      | 0   | 1    |
|4        | 0        | 1    | 1      | 0   | 0    |
|5        | 0        | 1    | 0      | 0   | 1    |
|2        | 0        | 0    | 1      | 1   | 0    |


### ASC Ordered

| Service | Critical | High | Medium | Low | None |
|---------|----------|------|--------|-----|------|
|2        | 0        | 0    | 1      | 1   | 0    |
|5        | 0        | 1    | 0      | 0   | 1    |
|4        | 0        | 1    | 1      | 0   | 0    |
|3        | 1        | 0    | 0      | 0   | 1    |
|1        | 1        | 0    | 0      | 1   | 0    |