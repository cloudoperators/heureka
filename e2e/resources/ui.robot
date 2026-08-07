# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    SeleniumLibrary

Resource   shadow.robot

*** Variables ***
${HEUREKA_UI_URL}         http://localhost:3000

${BROWSER}    headlessfirefox
#${BROWSER}    firefox

*** Keywords ***
Open browser to Heureka UI
    Open browser    ${HEUREKA_UI_URL}    ${BROWSER}
    Maximize browser window
    Test teardown append    Close Browser

Wait for heureka UI logo
    Wait until shadow element is visible    [data-testid="default-logo"]

Heureka UI is opened
    Open browser to Heureka UI
    Wait for Heureka UI logo
