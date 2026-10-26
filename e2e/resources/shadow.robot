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

Wait for shadow element text
    [Arguments]    ${selector}    ${timeout}=10s
    ${el}=    Wait for shadow element    ${selector}    ${timeout}
    ${text}=    Get text    ${el}
    Return from keyword    ${text}

Shadow element text should be equal
    [Arguments]    ${selector}    ${expected}    ${timeout}=10s
    ${actual_text}=    Wait for shadow element text    ${selector}    ${timeout}
    Should be equal    ${actual_text}    ${expected}

Shadow element text should contain
    [Arguments]    ${selector}    ${expected}    ${timeout}=10s
    ${actual_text}=    Wait for shadow element text    ${selector}    ${timeout}
    Should contain    ${actual_text}    ${expected}

Click shadow element
    [Arguments]    ${selector}    ${timeout}=10s
    ${el}=    Wait for shadow element    ${selector}    ${timeout}
    Execute Javascript    arguments[0].click();    ARGUMENTS    ${el}

Click shadow element with text
    [Arguments]    ${selector}    ${item_text}    ${timeout}=10s
    ${el}=    Get shadow element with text    ${selector}    ${item_text}    ${timeout}
    Click element     ${el}

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

Get all shadow elements
    [Arguments]    ${selector}
    ${elements}=    Execute javascript
    ...    return document.querySelector('[data-shadow-host="true"]').shadowRoot.querySelectorAll('${selector}');
    Run keyword if    ${{ $elements is None }}    Fail    Element ('${selector}') not found
    Return from keyword    ${elements}

Get shadow element with text
    [Arguments]    ${selector}    ${text}    ${timeout}=10s
    Wait for shadow element    ${selector}    ${timeout}
    ${element}=    Execute javascript
    ...    const root = document.querySelector('[data-shadow-host="true"]').shadowRoot;
    ...    const items = root.querySelectorAll('${selector}');
    ...    const target = Array.from(items).find(el => el.textContent.trim() === '${text}');
    ...    if (target) { return target; } else { throw new Error('Menu item not found'); }
    Run keyword if    ${{ $element is None }}    Fail    Element ('${selector}') with text: ('${text}') not found
    Return from keyword    ${element}

Wait for shadow element text to contain
    [Arguments]    ${selector}    ${expected}    ${timeout}=10s
    Wait until keyword succeeds    ${timeout}    500ms
    ...    Shadow element text should contain    ${selector}    ${expected}    ${timeout}
