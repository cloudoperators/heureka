# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

# Create a base layer with linkerd-await from a recent release.
FROM docker.io/curlimages/curl:latest as linkerd
ARG LINKERD_AWAIT_VERSION=v0.2.9
RUN curl -sSLo /tmp/linkerd-await https://github.com/linkerd/linkerd-await/releases/download/release%2F${LINKERD_AWAIT_VERSION}/linkerd-await-${LINKERD_AWAIT_VERSION}-amd64 && \
    chmod 755 /tmp/linkerd-await

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.3 AS builder
WORKDIR /go/src/github.com/cloudoperators/heureka/scanner/k8s-assets
ADD ./scanner/k8s-assets .
RUN CGO_ENABLED=0 go build -o /go/bin/heureka-scanner-k8s-assets main.go

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:nonroot
LABEL source_repository="https://github.com/cloudoperators/heureka"
LABEL source_folder="scanner/k8s-assets"
USER nonroot:nonroot

COPY --from=builder /go/bin/heureka-scanner-k8s-assets  /
COPY --from=linkerd /tmp/linkerd-await /linkerd-await

# Start linkerd
ENTRYPOINT ["/linkerd-await", "--shutdown", "--"]

# Start scanner
CMD ["/heureka-scanner-k8s-assets"]
