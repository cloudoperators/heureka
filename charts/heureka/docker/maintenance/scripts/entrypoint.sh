#!/bin/bash
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

set -e

# Validate environment variables
if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
    echo "Error: Required environment variables are not set"
    exit 1
fi

# Execute the maintenance script
echo "Starting database maintenance at $(date)"
mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < /scripts/maintenance.sql

echo "Database maintenance completed at $(date)"
