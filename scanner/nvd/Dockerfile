# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

# Create a base layer with linkerd-await from a recent release.
FROM docker.io/curlimages/curl:latest as linkerd
ARG LINKERD_AWAIT_VERSION=v0.2.9
RUN curl -sSLo /tmp/linkerd-await https://github.com/linkerd/linkerd-await/releases/download/release%2F${LINKERD_AWAIT_VERSION}/linkerd-await-${LINKERD_AWAIT_VERSION}-amd64 && \
    chmod 755 /tmp/linkerd-await

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.3 AS builder
COPY --from=linkerd /tmp/linkerd-await /linkerd-await

WORKDIR /go/src/github.com/cloudoperators/heureka/scanner/nvd
ADD ./scanner/nvd .
RUN CGO_ENABLED=0 go build -o /go/bin/heureka-scanner-nvd main.go

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:nonroot
LABEL source_repository="https://github.com/cloudoperators/heureka"
LABEL source_folder="scanner/nvd"
USER nonroot:nonroot

COPY --from=builder /go/bin/heureka-scanner-nvd  /

# Start linkerd
ENTRYPOINT ["/linkerd-await", "--shutdown", "--"]

# Start scanner
CMD ["/heureka-scanner-nvd"]
