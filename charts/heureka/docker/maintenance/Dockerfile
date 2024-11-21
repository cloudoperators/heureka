FROM alpine:3.20

# Install mysql-client and bash (for scripts)
RUN apk add --no-cache mysql-client bash

# Create non-root user
RUN addgroup -g 1000 dbuser && \
    adduser -u 1000 -G dbuser -s /bin/bash -D dbuser

# Create directory for scripts
WORKDIR /scripts
COPY scripts/* /scripts/

# Set permissions
RUN chown -R dbuser:dbuser /scripts && \
    chmod -R 755 /scripts

# Switch to non-root user
USER dbuser

# Entry point script
ENTRYPOINT ["/scripts/entrypoint.sh"]
