FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.2 AS builder

WORKDIR /go/src/github.com/cloudoperators/heureka
ADD . .
RUN go generate ./...
RUN CGO_ENABLED=0 go build -o /go/bin/heureka cmd/heureka/main.go

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:nonroot

LABEL source_repository="https://github.com/cloudoperators/heureka"
USER nonroot:nonroot
COPY --from=builder /go/bin/heureka /
ENTRYPOINT ["/heureka"]
