FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.3 AS builder

WORKDIR /go/src/github.com/cloudoperators/heureka/scanner/k8s-assets
ADD ../.. .
RUN CGO_ENABLED=0 go build -o /go/bin/heureka-scanner-k8s-assets main.go

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:nonroot

LABEL source_repository="https://github.com/cloudoperators/heureka"
LABEL source_folder="scanner/k8s-assets"
USER nonroot:nonroot

COPY --from=builder /go/bin/heureka-scanner-k8s-assets  /
ENTRYPOINT ["/heureka-scanner-k8s-assets"]
