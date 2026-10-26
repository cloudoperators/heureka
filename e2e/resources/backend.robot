# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    RequestsLibrary
Resource    graphql.robot

*** Variables ***
${HEUREKA_BACKEND_URL}                   http://localhost:80
${HEUREKA_BACKEND_GRAPHQL_ENDPOINT}      /query
${HEUREKA_BACKEND_MVREFRESH_ENDPOINT}    /internal/testing/mvrefresh

*** Keywords ***
Backend health request is sent
    GET    ${HEUREKA_BACKEND_URL}/health

Backend MVs refresh
     ${response}=    POST    ${HEUREKA_BACKEND_URL}${HEUREKA_BACKEND_MVREFRESH_ENDPOINT}
     Should Be Equal As Integers    ${response.status_code}    204

Backend invalid request is sent
    ${mutation_string}=    Catenate    SEPARATOR=\n
    ...    mutation {
    ...      addServiceToSupportGroup (
    ...        supportGroupId: "300"
    ...        serviceId: "600"
    ...      ) {
    ...        id
    ...        ccrn
    ...      }
    ...    }
    GraphQL mutation
    ...    mutation_string=${mutation_string}
    ...    mutation_name=addServiceToSupportGroup
    ...    support_group_id=invalidData
    ...    service_id=600
