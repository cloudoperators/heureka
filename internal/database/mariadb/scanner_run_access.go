// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

const defaultTimeoutForScannerRunLock = 10 * time.Second

type scannerRunAccess struct {
	lock chan struct{}
}

func newScannerRunAccess() *scannerRunAccess {
	return &scannerRunAccess{
		lock: make(chan struct{}, 1),
	}
}

func (sra *scannerRunAccess) Acquire(ctx context.Context) error {
	select {
	case sra.lock <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (sra *scannerRunAccess) Release() {
	<-sra.lock
}

func (s *SqlDatabase) acquireScannerRunAccess(t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	if err := s.scannerRunAcc.Acquire(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logrus.Error("Timed out waiting for operation lock")
		}

		return err
	}

	return nil
}

func (s *SqlDatabase) releaseScannerRunAccess() {
	s.scannerRunAcc.Release()
}
