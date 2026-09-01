FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.27 AS builder

WORKDIR /go/src/github.com/cloudoperators/heureka
ADD . .

# generate mock code files
RUN go install github.com/vektra/mockery/v3@v3.7.4
RUN mockery
# generate graphql code
RUN cd internal/api/graphql && go run github.com/99designs/gqlgen generate

RUN CGO_ENABLED=0 go build -o /go/bin/heureka cmd/heureka/main.go

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:nonroot

LABEL source_repository="https://github.com/cloudoperators/heureka"
USER nonroot:nonroot
COPY --from=builder /go/bin/heureka /
ENTRYPOINT ["/heureka"]
