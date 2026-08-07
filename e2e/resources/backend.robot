# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    RequestsLibrary

*** Variables ***
${HEUREKA_BACKEND_URL}                 http://localhost:80
${HEUREKA_BACKEND_GRAPHQL_ENDPOINT}    /query

*** Keywords ***
Backend health request is sent
    GET    ${HEUREKA_BACKEND_URL}/health
