# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    RequestsLibrary

*** Variables ***

*** Keywords ***
Get all user names
    [Documentation]    Iterates through Heureka pages using nextPageAfter to compile a flat list of names.
    Create Session     heureka_session    ${HEUREKA_BACKEND_URL}    verify=True

    ${master_name_list}=    Create list
    ${has_next}=            Set variable    ${TRUE}
    ${cursor}=              Set variable    ${NONE}
    
   ${query_string}=   Catenate    SEPARATOR=\n
    ...    query ($filter: UserFilter, $first: Int, $after: String) {
    ...        Users (
    ...            filter: $filter,
    ...            first: $first,
    ...            after: $after
    ...        ) {
    ...            totalCount
    ...            pageInfo {
    ...              hasNextPage
    ...             nextPageAfter
    ...            }
    ...            edges {
    ...                node {
    ...                    name
    ...                }
    ...            }
    ...        }
    ...    }


    WHILE    ${has_next}
        ${empty_list}=     Create List
        ${user_filter}=    Create dictionary    userName=${empty_list}
        
        # Pass the dynamic cursor token as the $after variable
        ${variables}=      Create dictionary    filter=${user_filter}    first=${10}    after=${cursor}
        ${payload}=        Create dictionary    query=${query_string}    variables=${variables}
        ${headers}=        Create dictionary    Content-Type=application/json    Accept=application/json
        
        ${response}=       POST on session      heureka_session    ${HEUREKA_BACKEND_GRAPHQL_ENDPOINT}    json=${payload}    headers=${headers}
        Status should be   200    ${response}
        ${json_res}=       Set variable         ${response.json()}
        
        # 2. FIXED PARSING LOCATIONS: Extract variables using the correct nextPageAfter map pointer
        ${has_next}=       Set variable         ${json_res['data']['Users']['pageInfo']['hasNextPage']}
        ${cursor}=         Set variable         ${json_res['data']['Users']['pageInfo']['nextPageAfter']}
        
        # Append this page's chunk straight to the master array tracking list
        ${edges}=          Set variable         ${json_res['data']['Users']['edges']}
        FOR    ${edge}    IN    @{edges}
            Append to list    ${master_name_list}    ${edge['node']['name']}
        END
    END
    
    RETURN    ${master_name_list}

Create service
	[Arguments]    ${ccrn}    ${domain}    ${region}
    Create Session     heureka_session    ${HEUREKA_BACKEND_URL}    verify=True

    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: ServiceInput!) {
    ...      createService (
    ...        input: $input
    ...      ) {
    ...        id
    ...        ccrn
    ...        domain
    ...        region
    ...      }
    ...    }

    ${service_input}=    Create dictionary    ccrn=${ccrn}    domain=${domain}    region=${region}

    ${variables}=    Create dictionary    input=${service_input}
    ${payload}=      Create dictionary    query=${mutation_string}    variables=${variables}
    ${headers}=      Create dictionary    Content-Type=application/json    Accept=application/json

    ${response}=    POST on session      heureka_session    ${HEUREKA_BACKEND_GRAPHQL_ENDPOINT}    json=${payload}    headers=${headers}
    Status should be   200    ${response}
    ${json_res}=       Set variable         ${response.json()}

    Return from keyword    ${json_res['data']['createService']['id']}
