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
${VULNERABILITY_MSG_SELECTOR}    div[class*="juno-stack"][class*="jn:flex"]:not(:has(button)):not(:has(input)) > div[class*="juno-stack"][class*="jn:flex"]
${SERVICE_MSG_SELECTOR}          div[class*="juno-datagrid-cell"][role*="gridcell"]
${FILTER_KEY_SELECTOR}           span[class*="pill-key"]
${FILTER_VALUE_SELECTOR}         span[class*="pill-value"]

*** Keywords ***
No vulnerabilities are found
    ${element}=    Wait for shadow element    ${VULNERABILITY_MSG_SELECTOR}    timeout=10s
    ${actual_text}=    Get text    ${element}
    Should contain    ${actual_text}    No vulnerabilities found!

No service matching criteria are found
    ${element}=    Wait for shadow element    ${SERVICE_MSG_SELECTOR}    timeout=10s
    ${actual_text}=    Get text    ${element}
    Should contain    ${actual_text}    No service found

Location should be changed to filter Services using SupportGroupCcrn
    Wait Until Location Contains    /services?f_supportGroupCcrn=containers    timeout=10s

Service tab with SupportGroupCcrn filter is visible
    ${key_element}=    Wait for shadow element    ${FILTER_KEY_SELECTOR}    timeout=10s
    ${key_text}=    Get text    ${key_element}
    Should contain    ${key_text}    supportGroupCcrn

    ${value_element}=    Wait for shadow element    ${FILTER_VALUE_SELECTOR}    timeout=10s
    ${value_text}=    Get text    ${value_element}
    Should contain    ${value_text}    containers

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

Heureka UI shows empty database on start screen
    Given Clear database
     When Heureka UI is opened
     Then Location should be changed to filter Services using SupportGroupCcrn
      And Service tab with SupportGroupCcrn filter is visible
      And No vulnerabilities are found
      And No service matching criteria are found
