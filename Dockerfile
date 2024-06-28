FROM golang:1.22.4 AS builder

WORKDIR /go/src/github.com/cloudoperators/heureka
ADD . .
RUN CGO_ENABLED=0 go build -o /go/bin/heureka cmd/heureka/main.go

FROM gcr.io/distroless/static-debian12:nonroot

LABEL source_repository="https://github.com/cloudoperators/heureka"
USER nonroot:nonroot
COPY --from=builder /go/bin/heureka /
ENTRYPOINT ["/heureka"]