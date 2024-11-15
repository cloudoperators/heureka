FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.3 AS builder

WORKDIR /go/src/github.com/cloudoperators/heureka/scanner/keppel

# Uncomment when scanner is available
ADD . .
RUN CGO_ENABLED=0 go build -o /go/bin/heureka-scanner-keppel main.go

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:nonroot

LABEL source_repository="https://github.com/cloudoperators/heureka"
USER nonroot:nonroot

# Uncomment when scanner is available
COPY --from=builder /go/bin/heureka-scanner-keppel  /heureka-scanner-keppel
ENTRYPOINT ["/heureka-scanner-keppel"]
