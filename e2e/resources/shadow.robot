# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    SeleniumLibrary

*** Keywords ***
Wait until shadow element is visible
    [Arguments]    ${selector}    ${timeout}=10s
    Wait until keyword succeeds    ${timeout}    500ms
    ...    Shadow element should exist    ${selector}

Wait for shadow element
    [Arguments]    ${selector}    ${timeout}=10s
    ${element}=    Wait until keyword succeeds    ${timeout}    500ms
    ...    Get shadow element    ${selector}
	Return from keyword    ${element}

Shadow element should exist
    [Arguments]    ${selector}

    ${found}=    Execute javascript
    ...    return document.querySelector('[data-shadow-host="true"]').shadowRoot.querySelector('${selector}') !== null

    Should Be True    ${found}

Get shadow element
    [Arguments]    ${selector}
    ${element}=    Execute javascript
    ...    return document.querySelector('[data-shadow-host="true"]').shadowRoot.querySelector('${selector}')
    Run keyword if    ${{ $element is None }}    Fail    Element ('${selector}') not found
    Return from keyword    ${element}
