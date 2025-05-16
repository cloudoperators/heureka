#!/usr/bin/bash
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0


sudo mariadb -pmariadb <<EOF
    -- Create a database and user for the application
    CREATE DATABASE IF NOT EXISTS heureka;
    
    -- Set root password
    ALTER USER 'root'@'localhost' IDENTIFIED BY 'mariadb';

    -- Create user vscode with password mariadb
    CREATE USER IF NOT EXISTS 'vscode'@'localhost' IDENTIFIED BY 'mariadb';
    GRANT ALL PRIVILEGES  ON *.*  TO 'vscode'@'localhost';
FLUSH PRIVILEGES;
EOF