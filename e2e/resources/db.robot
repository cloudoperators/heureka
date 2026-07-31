# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    DatabaseLibrary
Library    Process

*** Variables ***
${DB_HOST}              localhost
${DB_PORT}              3306
${DB_NAME}              heureka
${DB_USER_NAME}         my_username
${DB_USER_PASSWORD}     my_password
${DB_MIGRATIONS_DIR}    ${CURDIR}/../../internal/database/mariadb/migrations
${MIGRATE_TOOL}         migrate

*** Keywords ***
Connection to database is established
    Connect to database
    ...    pymysql
    ...    ${DB_NAME}
    ...    ${DB_USER_NAME}
    ...    ${DB_USER_PASSWORD}
    ...    ${DB_HOST}
    ...    ${DB_PORT}
    Test teardown append    Disconnect from database

Database migration dirty bit should be ${bitval}
    ${result}=    Query
    ...    SELECT version, dirty FROM schema_migrations
    Should not be empty    ${result}
    Should be equal as strings    ${result[0][1]}    ${bitval}

User table should contain only systemuser
    ${result}=    Query
    ...    SELECT user_name FROM user;
    Length should be    ${result}    1
    Should be equal    ${result[0][0]}    systemuser

All data tables should be empty
    ${query}=    Catenate
    ...    SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES
    ...    WHERE TABLE_TYPE = 'BASE TABLE' AND
    ...    TABLE_SCHEMA = 'heureka' AND
    ...    TABLE_NAME != 'schema_migrations' AND
    ...    TABLE_NAME != 'user' AND
    ...    TABLE_ROWS != 0;
    ${result}=    Query    ${query}
    Should be empty    ${result}

Feed database from gzip
    [Arguments]    ${gzip_path}    ${db_user}=${DB_USER_NAME}    ${db_password}=${DB_USER_PASSWORD}    ${db_name}=${DB_NAME}    ${db_host}=${DB_HOST}    ${db_port}=${DB_PORT}
    [Documentation]    Streams a .sql.gz file directly into the specified MariaDB database.

    # Construct the native shell pipeline command
    ${command}=    Catenate
    ...    set -o pipefail;
    ...    gunzip -c "${gzip_path}" |
    ...    mariadb -u"${db_user}" -P"${db_port}" -p"${db_password}" -h"${db_host}" "${db_name}"

    # shell=True is mandatory to allow the pipe (|) character to function
    ${result}=    Run process    ${command}    shell=True    stderr=STDOUT

    # Validate that the import executed successfully
    Log    ${result.stdout}
    Should be equal as integers    ${result.rc}    0
    ...    Database import failed with error: ${result.stdout}

Reload database schema
    Connect to database
    ...    pymysql
    ...    ${DB_NAME}
    ...    ${DB_USER_NAME}
    ...    ${DB_USER_PASSWORD}
    ...    ${DB_HOST}
    ...    ${DB_PORT}

    Execute sql string    DROP DATABASE IF EXISTS ${DB_NAME};
    Execute sql string    CREATE DATABASE ${DB_NAME};

    Disconnect from database

Run database up migrations
    [Arguments]    ${db_user}=${DB_USER_NAME}    ${db_password}=${DB_USER_PASSWORD}    ${db_name}=${DB_NAME}    ${db_host}=${DB_HOST}    ${db_port}=${DB_PORT}
    ${result}=    Run process
    ...    ${MIGRATE_TOOL}
    ...    -path
    ...    ${DB_MIGRATIONS_DIR}
    ...    -database
    ...    'mysql://${db_user}:${db_password}@tcp(${db_host}:${db_port})/${db_name}'
    ...    up
    ...    shell=True
    ...    stderr=STDOUT
    Should be equal as integers    ${result.rc}    0
    ...    Database up migration failed with error: ${result.stdout}

Run database down migrations
    [Arguments]    ${db_user}=${DB_USER_NAME}    ${db_password}=${DB_USER_PASSWORD}    ${db_name}=${DB_NAME}    ${db_host}=${DB_HOST}    ${db_port}=${DB_PORT}
    ${result}=    Run process
    ...    ${MIGRATE_TOOL}
    ...    -path
    ...    ${DB_MIGRATIONS_DIR}
    ...    -database
    ...    'mysql://${db_user}:${db_password}@tcp(${db_host}:${db_port})/${db_name}'
    ...    down
    ...    -all
    ...    shell=True
    ...    stderr=STDOUT
    Should be equal as integers    ${result.rc}    0
    ...    Database up migration failed with error: ${result.stdout}

Clear database
    Reload database schema
    Run database up migrations
