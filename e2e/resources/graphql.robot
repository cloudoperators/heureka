# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    RequestsLibrary

*** Variables ***
${ISSUE_TYPE_VULNERABILITY}           Vulnerability
${USER_TYPE_HUMAN}                    user
${COMPONENT_TYPE_CONTAINER_IMAGE}     containerImage
${COMPONENT_INSTANCE_TYPE_PROJECT}    Project
${ISSUE_MATCH_STATUS_VALUE_NEW}       new

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

Create object with input
    [Arguments]    ${mutation_string}    ${mutation_name}    &{input}
    ${variables}=    Create dictionary    input=${input}
    ${id}=    Create object    ${mutation_string}    ${mutation_name}    &{variables}
    Return from keyword    ${id}

Create object
    [Arguments]    ${mutation_string}    ${mutation_name}    &{variables}
    ${response}=    GraphQL mutation    ${mutation_string}    ${mutation_name}    &{variables}

    Status should be   200    ${response}
    ${json_res}=       Set variable         ${response.json()}

    Return from keyword    ${json_res['data']['${mutation_name}']['id']}

GraphQL mutation
    [Arguments]    ${mutation_string}    ${mutation_name}    &{variables}
    Create Session     heureka_session    ${HEUREKA_BACKEND_URL}    verify=True

    ${payload}=      Create dictionary    query=${mutation_string}    variables=${variables}
    ${headers}=      Create dictionary    Content-Type=application/json    Accept=application/json

    ${response}=    POST on session      heureka_session    ${HEUREKA_BACKEND_GRAPHQL_ENDPOINT}    json=${payload}    headers=${headers}
    Return from keyword    ${response}

Create service
    [Arguments]    ${ccrn}    ${domain}    ${region}

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

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createService
    ...    ccrn=${ccrn}
    ...    domain=${domain}
    ...    region=${region}
    Return from keyword    ${id}

Create supportGroup
    [Arguments]    ${ccrn}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: SupportGroupInput!) {
    ...      createSupportGroup (
    ...        input: $input
    ...      ) {
    ...        id
    ...        ccrn
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createSupportGroup
    ...    ccrn=${ccrn}
    Return from keyword    ${id}

Create issue
    [Arguments]    ${primaryName}    ${description}    ${type}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: IssueInput!) {
    ...      createIssue (
    ...        input: $input
    ...      ) {
    ...        id
    ...        primaryName
    ...        description
    ...        type
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createIssue
    ...    primaryName=${primaryName}
    ...    description=${description}
    ...    type=${type}
    Return from keyword    ${id}

Create user
    [Arguments]    ${uniqueUserId}    ${type}    ${name}    ${email}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: UserInput!) {
    ...      createUser (
    ...        input: $input
    ...      ) {
    ...        id
    ...        uniqueUserId
    ...        type
    ...        name
    ...        email
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createUser
    ...    uniqueUserId=${uniqueUserId}
    ...    type=${type}
    ...    name=${name}
    ...    email=${email}
    Return from keyword    ${id}

Create component
    [Arguments]    ${ccrn}    ${repository}    ${organization}    ${url}    ${type}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: ComponentInput!) {
    ...      createComponent (
    ...        input: $input
    ...      ) {
    ...        id
    ...        ccrn
    ...        repository
    ...        organization
    ...        url
    ...        type
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createComponent
    ...    ccrn=${ccrn}
    ...    repository=${repository}
    ...    organization=${organization}
    ...    url=${url}
    ...    type=${type}
    Return from keyword    ${id}

Create componentVersion
    [Arguments]    ${version}    ${componentId}    ${tag}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: ComponentVersionInput!) {
    ...      createComponentVersion (
    ...        input: $input
    ...      ) {
    ...        id
    ...        version
    ...        componentId
    ...        tag
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createComponentVersion
    ...    version=${version}
    ...    componentId=${componentId}
    ...    tag=${tag}
    Return from keyword    ${id}

Create componentInstance
    [Arguments]
    ...    ${ccrn}
    ...    ${region}
    ...    ${cluster}
    ...    ${namespace}
    ...    ${domain}
    ...    ${project}
    ...    ${pod}
    ...    ${container}
    ...    ${type}
    ...    ${count}
    ...    ${componentVersionId}
    ...    ${serviceId}
    ...    ${context}={}
    ...    ${parentId}=0
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: ComponentInstanceInput!) {
    ...      createComponentInstance (
    ...        input: $input
    ...      ) {
    ...        id
    ...        ccrn
    ...        region
    ...        cluster
    ...        namespace
    ...        domain
    ...        project
    ...        pod
    ...        container
    ...        type
    ...        context
    ...        count
    ...        componentVersionId
    ...        serviceId
    ...        parentId
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createComponentInstance
    ...    ccrn=${ccrn}
    ...    region=${region}
    ...    cluster=${cluster}
    ...    namespace=${namespace}
    ...    domain=${domain}
    ...    project=${project}
    ...    pod=${pod}
    ...    container=${container}
    ...    type=${type}
    ...    context=${context}
    ...    count=${count}
    ...    componentVersionId=${componentVersionId}
    ...    serviceId=${serviceId}
    ...    parentId=${parentId}
    Return from keyword    ${id}

Create issueRepository
    [Arguments]    ${name}    ${url}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: IssueRepositoryInput!) {
    ...      createIssueRepository (
    ...        input: $input
    ...      ) {
    ...        id
    ...        name
    ...        url
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createIssueRepository
    ...    name=${name}
    ...    url=${url}
    Return from keyword    ${id}

Create issueVariant
    [Arguments]    ${secondaryName}    ${description}    ${externalUrl}    ${severity}    ${issueRepositoryId}    ${issueId}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: IssueVariantInput!) {
    ...      createIssueVariant (
    ...        input: $input
    ...      ) {
    ...         id
    ...        secondaryName
    ...        description
    ...        externalUrl
    ...        severity {
    ...            value
    ...            score
    ...            cvss {
    ...                vector
    ...            }
    ...        }
    ...        issueRepositoryId
    ...        issueId
    ...      }
    ...    }

    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createIssueVariant
    ...    secondaryName=${secondaryName}
    ...    description=${description}
    ...    externalUrl=${externalUrl}
    ...    severity=${severity}
    ...    issueRepositoryId=${issueRepositoryId}
    ...    issueId=${issueId}
    Return from keyword    ${id}

Create issueMatch
    [Arguments]    ${status}    ${remediationDate}    ${discoveryDate}    ${targetRemediationDate}    ${componentInstanceId}    ${issueId}    ${userId}
    ${mutation_string}=   Catenate    SEPARATOR=\n
    ...    mutation ($input: IssueMatchInput!) {
    ...      createIssueMatch (
    ...        input: $input
    ...      ) {
    ...        id
    ...        status
    ...        remediationDate
    ...        discoveryDate
    ...        targetRemediationDate
    ...        componentInstanceId
    ...        issueId
    ...        userId
    ...        severity {
    ...          value
    ...          score
    ...          cvss {
    ...            vector
    ...          }
    ...        }
    ...      }
    ...    }
    ${id}=    Create object with input
    ...    mutation_string=${mutation_string}
    ...    mutation_name=createIssueMatch
    ...    status=${status}
    ...    remediationDate=${remediationDate}
    ...    discoveryDate=${discoveryDate}
    ...    targetRemediationDate=${targetRemediationDate}
    ...    componentInstanceId=${componentInstanceId}
    ...    issueId=${issueId}
    ...    userId=${userId}
    Return from keyword    ${id}

Add service to supportGroup
    [Arguments]    ${supportGroupId}    ${serviceId}
    ${mutation_string}=    Catenate    SEPARATOR=\n
    ...    mutation {
    ...      addServiceToSupportGroup (
    ...        supportGroupId: "${supportGroupId}"
    ...        serviceId: "${serviceId}"
    ...      ) {
    ...        id
    ...      }
    ...    }

    ${id}=    Create object
    ...    mutation_string=${mutation_string}
    ...    mutation_name=addServiceToSupportGroup
    ...    support_group_id=${supportGroupId}
    ...    service_id=${serviceId}
    Return from keyword    ${id}

Add componentVersion to issue
    [Arguments]    ${issueId}    ${componentVersionId}
    ${mutation_string}=    Catenate    SEPARATOR=\n
    ...    mutation {
    ...      addComponentVersionToIssue (
    ...        issueId: "${issueId}"
    ...        componentVersionId: "${componentVersionId}"
    ...      ) {
    ...        id
    ...      }
    ...    }

    ${id}=    Create object
    ...    mutation_string=${mutation_string}
    ...    mutation_name=addComponentVersionToIssue
    ...    issue_id=${issueId}
    ...    component_version_id=${componentVersionId}
    Return from keyword    ${id}
