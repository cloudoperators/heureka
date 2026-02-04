package test

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/types"
)

// CONTAIN ALL

func ContainAll(expected interface{}) types.GomegaMatcher {
	return &containAllMatcher{
		expected: expected,
	}
}

type containAllMatcher struct {
	expected interface{}
	missing  []interface{}
}

func (m *containAllMatcher) Match(actual interface{}) (bool, error) {
	actualVal := reflect.ValueOf(actual)
	expectedVal := reflect.ValueOf(m.expected)

	if actualVal.Kind() != reflect.Slice && actualVal.Kind() != reflect.Array {
		return false, fmt.Errorf("ContainAll matcher expects a slice/array as actual")
	}
	if expectedVal.Kind() != reflect.Slice && expectedVal.Kind() != reflect.Array {
		return false, fmt.Errorf("ContainAll matcher expects a slice/array as expected")
	}

	m.missing = nil

	for i := 0; i < expectedVal.Len(); i++ {
		exp := expectedVal.Index(i).Interface()

		found := false
		for j := 0; j < actualVal.Len(); j++ {
			if reflect.DeepEqual(actualVal.Index(j).Interface(), exp) {
				found = true
				break
			}
		}

		if !found {
			m.missing = append(m.missing, exp)
		}
	}

	return len(m.missing) == 0, nil
}

func (m *containAllMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n%v\nto contain all elements\n%v\nmissing: %v",
		actual,
		m.expected,
		m.missing,
	)
}

func (m *containAllMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n%v\nnot to contain all elements\n%v",
		actual,
		m.expected,
	)
}

// CONTAIN NONE

func ContainNone(forbidden interface{}) types.GomegaMatcher {
	return &containNoneMatcher{
		forbidden: forbidden,
	}
}

type containNoneMatcher struct {
	forbidden interface{}
	found     []interface{}
}

func (m *containNoneMatcher) Match(actual interface{}) (bool, error) {
	actualVal := reflect.ValueOf(actual)
	forbiddenVal := reflect.ValueOf(m.forbidden)

	if actualVal.Kind() != reflect.Slice && actualVal.Kind() != reflect.Array {
		return false, fmt.Errorf("ContainNone matcher expects slice/array as actual")
	}
	if forbiddenVal.Kind() != reflect.Slice && forbiddenVal.Kind() != reflect.Array {
		return false, fmt.Errorf("ContainNone matcher expects slice/array as forbidden")
	}

	m.found = nil

	for i := 0; i < forbiddenVal.Len(); i++ {
		f := forbiddenVal.Index(i).Interface()

		for j := 0; j < actualVal.Len(); j++ {
			if reflect.DeepEqual(actualVal.Index(j).Interface(), f) {
				m.found = append(m.found, f)
				break
			}
		}
	}

	return len(m.found) == 0, nil
}

func (m *containNoneMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n%v\nnot to contain any of\n%v\nbut found: %v",
		actual,
		m.forbidden,
		m.found,
	)
}

func (m *containNoneMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n%v\nto contain at least one of\n%v",
		actual,
		m.forbidden,
	)
}
