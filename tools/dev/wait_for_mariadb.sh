#!/bin/sh

# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

HOST=${DB_ADDRESS}
PORT=${DB_PORT}

while ! nc -z "$HOST" "$PORT" >/dev/null 2>&1; do
  echo "Waiting for MariaDB to be ready..."
  sleep 1
done

echo "MariaDB is ready!"