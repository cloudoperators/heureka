// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"strconv"

	"golang.org/x/exp/constraints"
)

// FindMaxBy is getting a Maximum Value from a given slice using a callback function that is returning the value of type
// constraints.Ordered used to do the comparison
func FindMaxBy[E any, T constraints.Ordered](s []E, fn func(val E) (T, error)) (*E, error) {
	var max E

	if len(s) > 0 {
		max = s[0]
	}

	for i := 0; i < len(s); i++ {

		v, err := fn(s[i])
		if err != nil {
			return nil, errors.New("Error in FindMaxBy Callback function")
		}

		m, err := fn(max)
		if err != nil {
			return nil, errors.New("Error in FindMaxBy Callback function")
		}
		if v > m {
			max = s[i]
		}
	}

	return &max, nil
}

func ConvertStrToIntSlice(slice []*string) ([]*int64, error) {
	var result []*int64
	for _, str := range slice {
		if str != nil {
			val, err := strconv.ParseInt(*str, 10, 64)
			if err != nil {
				return nil, err
			}
			result = append(result, &val)
		} else {
			result = append(result, nil)
		}
	}
	return result, nil
}
