// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import (
	"fmt"
	"sync"
	"time"
)

type Executor struct {
	wg   sync.WaitGroup
	done chan struct{}
}

func Execute(fn func(), n int) Executor {
	e := Executor{}
	e.done = make(chan struct{})
	e.wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer e.wg.Done()
			fn()
		}()
	}

	go func() {
		e.wg.Wait()
		close(e.done)
	}()
	return e
}

func Wait(e Executor, timeout time.Duration) error {
	select {
	case <-e.done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for jobs to finish")
	}
}
