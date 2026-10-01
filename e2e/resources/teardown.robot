# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
# SPDX-License-Identifier: Apache-2.0

*** Settings ***
Library    Collections

*** Keywords ***
Test teardown init
    [Documentation]    Initialize a list of teardown actions which can be executed by 'Test teardown run' automatically as long as it is put into Test teardown field. Run this init in setup or case steps
    Variable should not exist    ${test_teardown_stack}    Test teardown init: test_teardown_stack already defined!!!
    @{test_teardown_stack}    Create list
    Set test variable    ${test_teardown_stack}

Test teardown run
    [Documentation]    Get list objects from list '*${test_teardown_stack}*', and then execute these list objects (keyword and its arguments) one by one. Failure will be ignored so all list object can be executed.
    ${status}    Run keyword and return status    Variable should exist    ${test_teardown_stack}    Test teardown run: test_teardown_stack not defined!!! Run 'Test teardown init' in setup or case.
    Return from keyword if    ${status} != True
    Log    ============== Test teardown run start ===============
    FOR    ${item}    IN    @{test_teardown_stack}
        ${status}    ${return_value}    Run keyword and ignore error    @{item}
        Log    ${item[0]} status:::${status}, return:::${return_value}
    END
    Log    ============== Test teardown run end ===============

Test teardown append
    [Arguments]    @{keyword_and_args}
    [Documentation]    *@{keyword_and_args}*: the list of keyword and its arguments
    Variable should exist    ${test_teardown_stack}    Test teardown add: test_teardown_stack not defined!!! Run 'Test teardown init' in setup or case.
    @{kw_args}    Create list    @{keyword_and_args}
    Append To List    ${test_teardown_stack}    ${kw_args}

Test teardown add head
    [Arguments]    @{keyword_and_args}
    [Documentation]    *@{keyword_and_args}*: the list of keyword and its arguments
    Variable should exist    ${test_teardown_stack}    Test teardown add head: test_teardown_stack not defined!!! Run 'Test teardown init' in setup or case.
    @{kw_args}    Create list    @{keyword_and_args}
    Insert into list    ${test_teardown_stack}    0    ${kw_args}

Test teardown insert
    [Arguments]    ${insert_position}=0    @{keyword_and_args}
    [Documentation]    *${insert_position}*: like 0 (the first), -1(the reversed second), -2(the reversed third), and so on.
    ...
    ...    *@{keyword_and_args}*: the list of keyword and its arguments
    Variable should exist    ${test_teardown_stack}    Test teardown insert: test_teardown_stack not defined!!! Run 'Test teardown init' in setup or case.
    @{kw_args}    Create list    @{keyword_and_args}
    Insert into list    ${test_teardown_stack}    ${insert_position}    ${kw_args}

Keyword teardown init
    [Documentation]    Initialize a list of teardown actions which can be executed by 'Keyword Teardown Run' automatically as long as it is put into Keyword's Teardown field. Run this init in keyword steps
    #Variable should not exist    ${keyword_teardown_queue}    Keyword teardown init: keyword_teardown_stack already defined!!!
    @{keyword_teardown_queue}    Create list
    Set test variable    ${keyword_teardown_queue}

Keyword teardown run
    [Documentation]    Get list objects from list '*${keyword_teardown_stack}*', and then execute these list objects (keyword and its arguments) one by one. Failure will be ignored so all list object can be executed.
    ${status}    Run keyword and return status    Variable should exist    ${keyword_teardown_queue}    Keyword teardown Run: keyword_teardown_queue not defined!!! Run 'Keyword teardown init' in setup or case.
    Return from keyword if    ${status} != True
    Log    ============== Test teardown run start ===============
    FOR    ${item}    IN    @{keyword_teardown_queue}
        ${status}    ${return_value}    Run keyword and ignore error    @{item}
        Log    ${item[0]} status:::${status}, return:::${return_value}
	END
    Log    ============== Test teardown run end ===============
    @{keyword_teardown_queue}    Create list

Keyword teardown add head
    [Arguments]    @{keyword_and_args}
    [Documentation]    *@{keyword_and_args}*: the list of keyword and its arguments
    Variable should exist    ${keyword_teardown_queue}    Keyword teardown add head: keyword_teardown_queue \ not defined!!! Run 'Keyword teardown init' in setup or case.
    @{kw_args}    Create list    @{keyword_and_args}
    Insert Into List    ${keyword_teardown_queue}    0    ${kw_args}

Keyword teardown append
    [Arguments]    @{keyword_and_args}
    [Documentation]    *@{keyword_and_args}*: the list of keyword and its arguments
    Variable should exist    ${keyword_teardown_queue}    Keyword teardown append: keyword_teardown_queue not defined!!! Run 'Keyword teardown init' in setup or case.
    @{kw_args}    Create list    @{keyword_and_args}
    Append To List    ${keyword_teardown_queue}    ${kw_args}

Keyword teardown insert
    [Arguments]    ${insert_position}=0    @{keyword_and_args}
    [Documentation]    *${insert_position}*: like 0 (the first), -1(the reversed second), -2(the reversed third), and so on.
    ...
    ...    *@{keyword_and_args}*: the list of keyword and its arguments
    Variable should exist    ${keyword_teardown_queue}    Keyword teardown insert: keyword_teardown_queue \ not defined!!! Run 'Keyword teardown init' in setup or case.
    @{kw_args}    Create list    @{keyword_and_args}
    Insert into list    ${keyword_teardown_queue}    ${insert_position}    ${kw_args}
