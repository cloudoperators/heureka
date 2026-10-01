# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    RequestsLibrary

Resource   ../resources/backend.robot
Resource   ../resources/db.robot
Resource   ../resources/graphql.robot
Resource   ../resources/teardown.robot
Resource   ../resources/ui.robot

Force tags        smoke

Test setup       Test teardown init
Test teardown    Test teardown run

*** Variables ***
${LIST_ROW_SELECTOR}                                           div[class*="juno-datagrid"][role="grid"] > div[class*="datagrid-row"][role="row"]:not(:has(div[class*="juno-datagrid-cell"][role="columnheader"]))
${LIST_CELL_SELECTOR}                                          ${LIST_ROW_SELECTOR} > div[class*="juno-datagrid-cell"][role="gridcell"]
${VULNERABILITY_COUNT_SELECTOR}                                div[class*="font-bold mr-2"]
${FILTER_KEY_SELECTOR}                                         span[class*="pill-key"]
${FILTER_VALUE_SELECTOR}                                       span[class*="pill-value"]
${IMAGE_COUNT_FOR_SERVICE_SELECTOR}                            span[class*="ml-1"]
${IMAGE_LIST_FOR_SERVICE_SELECTOR}                             ${LIST_CELL_SELECTOR} > div[class*="juno-stack"] > div[class*="juno-stack"] > span
${HEADING_SELECTOR}                                            h1[class*="juno-content-heading"]
${VULNERABILITY_LIST_SELECTOR}                                 ${LIST_CELL_SELECTOR} > div[class*="juno-stack"] > span
${ACTION_BUTTON_ON_VULNERABLE_SELECTOR}                        span:has(> svg[class*="juno-icon juno-icon-moreVert"])
${ACTION_ITEM_SELECTOR}                                        div[class*="juno-popupmenu-item"]
${REMEDIATION_CHANGE_SEVERITY_SEVERITY_SELECTOR}               div > button[class*="juno-select-toggle"][aria-label="New Severity"]
${SEVERITY_ITEM_SELECTOR}                                      li[class*="juno-select-option"] > span
${REMEDIATION_CHANGE_SEVERITY_EXPIRATION_DATE_SELECTOR}        div:has(> div > input[id*="juno-datetimepicker"][class*="juno-datetimepicker-input"])
${REMEDIATION_CHANGE_SEVERITY_CALENDAR_NEXT_MONTH_SELECTOR}    span[class*="flatpickr-next-month"]:has(> svg)
${REMEDIATION_CHANGE_SEVERITY_CALENDAR_DAY_SELECTOR}           span[class*="flatpickr-day"]
${REMEDIATION_CHANGE_SEVERITY_DESCRIPTION_SELECTOR}            textarea[id*="juno-textarea"][class*="juno-textarea"]
${REMEDIATION_CHANGE_SEVERITY_SUBMIT_SELECTOR}                 button[class*="juno-button"][title="Change Severity"]

${TEST_DATA}    ${{ { 'Test': { 'vulnerabilityPrimaryName': 'CVE-2026-12345' } } }}

*** Keywords ***
No vulnerabilities are found
    Shadow element text should contain    ${LIST_CELL_SELECTOR}    No vulnerabilities found!

No service matching criteria are found
    Shadow element text should be equal    ${LIST_CELL_SELECTOR}    No service found

Location should be changed to filter Services using SupportGroupCcrn
    Wait until location contains    /services?f_supportGroupCcrn=containers

Service tab with SupportGroupCcrn filter is visible
    Shadow element text should be equal    ${FILTER_KEY_SELECTOR}    supportGroupCcrn
    Shadow element text should be equal    ${FILTER_VALUE_SELECTOR}    containers

Get '${alias}' service ccrn
    Return from keyword    ${alias}ServiceCcrn

Get '${alias}' component repository
    Return from keyword    ${alias}ComponentRepository

Get '${alias}' vulnerability primary name
    ${alias_data}=    Get from dictionary    ${TEST_DATA}    ${alias}
    ${vulnerability_primary_name}=    Get from dictionary    ${alias_data}    vulnerabilityPrimaryName
    Return from keyword    ${vulnerability_primary_name}

'${alias}' vulnerability is created
    ${supportGroupId}=    Create supportGroup    containers
    ${vulnerabilityPrimaryName}=    Get '${alias}' vulnerability primary name
    ${issueId}=    Create issue    primaryName=${vulnerabilityPrimaryName}    description=${alias}Description    type=${ISSUE_TYPE_VULNERABILITY}
    ${issueRepositoryId}=    Create issueRepository    name=${alias}IRName    url=http://${alias}.url.co
    ${severity}=    Create dictionary    rating=High
    Create issueVariant
    ...    secondaryName=${alias}SecondaryName
    ...    description=${alias}Description
    ...    externalUrl=http://${alias}.external.url
    ...    severity=${severity}
    ...    issueRepositoryId=${issueRepositoryId}
    ...    issueId=${issueId}

    ${componentRepository}=    Get '${alias}' component repository
    ${componentId}=           Create component
    ...        ccrn=${alias}ComponentCcrn
    ...        repository=${componentRepository}
    ...        organization=${alias}ComponentOrganization
    ...        url=https://${alias}.component.url.com
    ...        type=${COMPONENT_TYPE_CONTAINER_IMAGE}
    ${componentVersionId}=    Create componentVersion
    ...        version=sha256:9928d59477fb27e511f545aba976b4f5196ac4ca9779f26fc50fc7228986f013
    ...        componentId=${componentId}
    ...        tag=v0.0.1
    ${serviceCcrn}=    Get '${alias}' service ccrn
    ${serviceId}=             Create service    ccrn=${serviceCcrn}    domain=${alias}ServiceDomain    region=${alias}ServiceRegion
    ${componentInstanceId}=    Create componentInstance
    ...    ccrn=${alias}ComponentInstanceCcrn
    ...    region=${alias}Region
    ...    cluster=${alias}Cluster
    ...    namespace=${alias}Namespace
    ...    domain=${alias}Domain
    ...    project=${alias}Project
    ...    pod=${alias}Pod
    ...    container=${alias}Container
    ...    type=${COMPONENT_INSTANCE_TYPE_PROJECT}
    ...    count=1
    ...    componentVersionId=${componentVersionId}
    ...    serviceId=${serviceId}

    ${userId}=    Create user    uniqueUserId=U0000001    type=${USER_TYPE_HUMAN}    name=${alias}UserName    email=${alias}.user@dev.null

    Create issueMatch
    ...    status=${ISSUE_MATCH_STATUS_VALUE_NEW}
    ...    remediationDate=${None}
    ...    discoveryDate=2026-09-01T18:30:00Z
    ...    targetRemediationDate=${None}
    ...    componentInstanceId=${componentInstanceId}
    ...    issueId=${issueId}
    ...    userId=${userId}

    Add componentVersion to issue    issueId=${issueId}    componentVersionId=${componentVersionId}
    Add service to supportGroup    supportGroupId=${supportGroupId}    serviceId=${serviceId}

    Backend MVs refresh

All vulnerability count for selected services is ${count}
    ${element}=    Wait for shadow element    ${VULNERABILITY_COUNT_SELECTOR}
    ${actual_text}=    Get text    ${element}
    Should be equal as integers    ${actual_text}    ${count}

Image count for service is ${count}
    ${element}=    Wait for shadow element    ${IMAGE_COUNT_FOR_SERVICE_SELECTOR}
    ${actual_text}=    Get text    ${element}
    Should be equal as integers    ${actual_text}    ${count}

Find '${alias}' vulnerability row in the list
    ${vulnerabilityPrimaryName}=    Get '${alias}' vulnerability primary name
    ${rows}=    Get all shadow elements    ${LIST_ROW_SELECTOR}
    FOR    ${row}    IN    @{rows}
        ${ce}=    Call Method    ${row}    find_element    by=css selector    value=div[class*="juno-datagrid-cell"][role="gridcell"] > div[class*="juno-stack"] > span
        ${text}=    Get text    ${ce}
        Return from keyword if    '${text}' == '${vulnerabilityPrimaryName}'    ${row}
    END
    Fail    '${alias}' vulnerability row not found in the list

Click action menu '${menu_item}' for '${alias}' vulnerability
    ${vulnerability_row}=    Find '${alias}' vulnerability row in the list
    ${action_button}=    Call Method    ${vulnerability_row}    find_element    by=css selector    value=span:has(> svg[class*="juno-icon juno-icon-moreVert"])
    Click element    ${action_button}
    Click shadow element with text    ${ACTION_ITEM_SELECTOR}    ${menu_item}

'${alias}' vulnerability is visible
    All vulnerability count for selected services is 1
    Click on '${alias}' vulnerable service
    Image count for service is 1
    Click on '${alias}' image for service
    '${alias}' vulnerability is visible on the list of active vulnerabilities

Change severity remediation is issued on the '${alias}' vulnerability
    Click action menu 'Change Severity' for '${alias}' vulnerability
    Remediation change severity ticket for '${alias}' vulnerability should be visible
    Fill and submit change severity remediation    Low

Click on '${alias}' vulnerable service
    ${expected}=    Get '${alias}' service ccrn
    Wait for shadow element text to contain    ${LIST_CELL_SELECTOR}    ${expected}
    Click shadow element    ${LIST_CELL_SELECTOR}
    Wait for shadow element text to contain    ${HEADING_SELECTOR}    Service
    ${serviceCcrn}=    Get '${alias}' service ccrn
    Shadow element text should contain    ${HEADING_SELECTOR}    ${serviceCcrn}

Click on '${alias}' image for service
    ${expected}=    Get '${alias}' component repository
    Shadow element text should be equal    ${IMAGE_LIST_FOR_SERVICE_SELECTOR}    ${expected}
    Click shadow element    ${IMAGE_LIST_FOR_SERVICE_SELECTOR}
    Wait for shadow element text to contain    ${HEADING_SELECTOR}    Image
    ${image}=    Get '${alias}' component repository
    Shadow element text should contain    ${HEADING_SELECTOR}    ${image}

'${alias}' vulnerability is visible on the list of active vulnerabilities
    ${vulnerabilityPrimaryName}=    Get '${alias}' vulnerability primary name
    Get shadow element with text    ${VULNERABILITY_LIST_SELECTOR}    ${vulnerabilityPrimaryName}

Remediation change severity ticket for '${alias}' vulnerability should be visible
    ${expected_vulnerability_name}=    Get '${alias}' vulnerability primary name
    ${vulnerability}=    Get shadow element with text    div > strong    Vulnerability:
    ${text}=    Get WebElement parent text    ${vulnerability}
    Should be equal    '${text}'    'Vulnerability: ${expected_vulnerability_name}'

    ${expected_service}=    Get '${alias}' service ccrn
    ${service}=    Get shadow element with text    div > strong    Service:
    ${text}=    Get WebElement parent text    ${service}
    Should be equal    '${text}'    'Service: ${expected_service}'

    ${expected_image}=    Get '${alias}' component repository
    ${image}=    Get shadow element with text    div > strong    Image:
    ${text}=    Get WebElement parent text    ${image}
    Should be equal    '${text}'    'Image: ${expected_image}'

    ${current_severity}=    Get shadow element with text    div > strong    Current Severity:
    ${text}=    Get WebElement parent text    ${current_severity}
    Should be equal    '${text}'    'Current Severity: High'

Get WebElement parent text
    [Arguments]    ${web_element}
    ${parent_element}=    Call Method    ${web_element}    find_element    by=xpath    value=parent::*
    ${text}=    Get text    ${parent_element}
    Return from keyword    ${text}

Fill and submit change severity remediation
    [Arguments]    ${new_severity}
    Click shadow element    ${REMEDIATION_CHANGE_SEVERITY_SEVERITY_SELECTOR}
    Click shadow element with text    ${SEVERITY_ITEM_SELECTOR}    ${new_severity}
    Pick date in the future for remediation change severity expiration date
    Set remediation change severity description    someDescription
    Click shadow element    ${REMEDIATION_CHANGE_SEVERITY_SUBMIT_SELECTOR}

Pick date in the future for remediation change severity expiration date
    ${calendar_button}=    Get shadow element    ${REMEDIATION_CHANGE_SEVERITY_EXPIRATION_DATE_SELECTOR}
    Click element    ${calendar_button}
    Click shadow element    ${REMEDIATION_CHANGE_SEVERITY_CALENDAR_NEXT_MONTH_SELECTOR}
    Click shadow element with text    ${REMEDIATION_CHANGE_SEVERITY_CALENDAR_DAY_SELECTOR}    25

Set remediation change severity description
    [Arguments]    ${description_text}
    ${description_text_field_element}=    Get shadow element    ${REMEDIATION_CHANGE_SEVERITY_DESCRIPTION_SELECTOR}
    Input text   ${description_text_field_element}    ${description_text}

'${alias}' vulnerability is not visible on the list of active vulnerabilities
    Wait for shadow element text to contain    ${LIST_CELL_SELECTOR}    No vulnerabilities found!

'${alias}' vulnerability is visible on the list of remediated vulnerabilities
    Click shadow element with text    li[class*="juno-tab"][role="tab"]    Remediated Vulnerabilities
    ${vulnerabilityPrimaryName}=    Get '${alias}' vulnerability primary name
    Get shadow element with text    ${VULNERABILITY_LIST_SELECTOR}    ${vulnerabilityPrimaryName}

*** Test Cases ***
Heureka UI is operational
    Given Open browser to Heureka UI
     When Wait for Heureka UI logo
     Then Title should be    Heureka

Heureka Backend is healthy
     When Backend health request is sent
     Then Status should be    200

Database is available
     When Connection to database is established
     Then Database migration dirty bit should be 0

Database schema is empty
     When Connection to database is established
     Then User table should contain only systemuser
      And All data tables should be empty

Heureka UI shows empty data on start screen
    [Tags]    yyy
    Given Database is cleared
     When Heureka UI is opened
     Then Location should be changed to filter Services using SupportGroupCcrn
      And Service tab with SupportGroupCcrn filter is visible
      And No vulnerabilities are found
      And No service matching criteria are found

Heureka Backend responds with error on invalid graphql query
     When Backend invalid request is sent
     Then Status should be    400

Heureka UI shows service with vulnerability
    Given Database is cleared
      And 'Test' vulnerability is created
     When Heureka UI is opened
     Then 'Test' vulnerability is visible

Heureka UI shows remediated vulnerability
    [Tags]    xxx
    Given Database is cleared
      And 'Test' vulnerability is created
      And Heureka UI is opened
      And 'Test' vulnerability is visible
     When Change severity remediation is issued on the 'Test' vulnerability
     Then 'Test' vulnerability is not visible on the list of active vulnerabilities
      And 'Test' vulnerability is visible on the list of remediated vulnerabilities
