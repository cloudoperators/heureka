# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

FROM node:20-bullseye AS builder

# Install jq and pnpm
RUN apt-get update && apt-get install -y jq git && \
    npm install -g pnpm turbo && \
    rm -rf /var/lib/apt/lists/*

# Set build-time environment variables
ARG UI_GIT_DIR
ARG UI_GIT_REPO

WORKDIR /app

# Clone or update repo
RUN if [ -d "$UI_GIT_DIR" ]; then \
        cd "$UI_GIT_DIR" && git pull -r && cd ../ ; \
    else \
        git clone "$UI_GIT_REPO" "$UI_GIT_DIR" ; \
    fi

# Modify JSON config
RUN cd "$UI_GIT_DIR" && \
    jq '.apiEndpoint = "http://localhost:80/query" | .initialFilters.support_group = ["containers"]' \
       apps/heureka/appProps.template.json > apps/heureka/appProps.json

# Build
RUN cd "$UI_GIT_DIR" && \
    pnpm run clean:cache && \
    pnpm run reinstall && \
    pnpm run build

# Runtime stage
FROM node:20-bullseye
RUN npm install -g pnpm turbo

# Set Runtime environment variables
ARG UI_GIT_DIR

# Copy build artifacts
WORKDIR /app
COPY --from=builder /app/$UI_GIT_DIR  /app

# Default command to run the dev server
CMD ["pnpx", "turbo", "dev", "--filter", "@cloudoperators/juno-app-heureka"]
