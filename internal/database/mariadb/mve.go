// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

type MvEngine struct {
	scheduler         *gocron.Scheduler
	firstRunDone      chan struct{}
	once              sync.Once
	cfg               util.Config
	triggerMu         sync.Mutex
	refreshMu         sync.Mutex
	running           bool
	pending           bool
	OnRefreshComplete func()
}

func NewMvEngine(cfg util.Config) *MvEngine {
	return &MvEngine{
		scheduler:    gocron.NewScheduler(time.UTC),
		firstRunDone: make(chan struct{}),
		cfg:          cfg,
	}
}

func TriggerMVE(cfg util.Config) error {
	db, err := NewDb(cfg)
	if err != nil {
		return fmt.Errorf("error while Creating Db: %w", err)
	}

	return runInBackground(db, MVProcedures).Wait()
}

// runFullRefresh serializes the whole build-and-swap cycle (scheduled run and async
// trigger alike) under refreshMu, so no two full refreshes collide on the shared
// _tmp/_old tables and no swap lands on top of a concurrent targeted write.
func (mve *MvEngine) runFullRefresh() error {
	mve.refreshMu.Lock()
	defer mve.refreshMu.Unlock()

	return TriggerMVE(mve.cfg)
}

func (mve *MvEngine) Start() {
	periodMinutes := mve.cfg.DBMvCalcPeriodMinutes
	if periodMinutes <= 0 {
		periodMinutes = 200
	}

	logrus.Debugf("MVE scheduling period set to %d minutes", periodMinutes)

	_, err := mve.scheduler.Every(periodMinutes).Minutes().SingletonMode().Do(func() {
		err := mve.runFullRefresh()
		if err != nil {
			logrus.WithError(err).Error("MVE Trigger error")
		}

		mve.once.Do(func() {
			close(mve.firstRunDone)
		})
	})
	if err != nil {
		logrus.WithError(err).Error("MVE Do() error")
		mve.once.Do(func() {
			close(mve.firstRunDone)
		})
	}

	mve.scheduler.StartAsync()
}

func (mve *MvEngine) Stop() {
	mve.scheduler.Clear()
	// The following method is not advisory as it may hang for a long time:
	// mve.scheduler.Stop()
}

func (mve *MvEngine) TriggerAsync() {
	mve.triggerMu.Lock()
	mve.pending = true

	if mve.running {
		mve.triggerMu.Unlock()
		return
	}

	mve.running = true
	mve.triggerMu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("MVE async trigger panicked: %v", r)
			}

			mve.triggerMu.Lock()
			mve.running = false
			pending := mve.pending
			mve.triggerMu.Unlock()

			if pending {
				mve.TriggerAsync()
			}
		}()

		for {
			mve.triggerMu.Lock()
			if !mve.pending {
				mve.triggerMu.Unlock()

				return
			}

			mve.pending = false
			mve.triggerMu.Unlock()

			func() {
				defer func() {
					if r := recover(); r != nil {
						logrus.Errorf("MVE async trigger panicked: %v", r)
					}
				}()

				if err := mve.runFullRefresh(); err != nil {
					logrus.WithError(err).Error("MVE async trigger error")
					return
				}

				if cb := mve.OnRefreshComplete; cb != nil {
					cb()
				}
			}()
		}
	}()
}

func (mve *MvEngine) WaitForFirstRun() {
	<-mve.firstRunDone
}

// TargetedVulnerabilityRefresh recomputes only the mvVulnerabilityList and
// mvVulnerabilityService rows affected by a remediation on (serviceID, issueID), so
// a rescore's new severity shows up without waiting for a full MV refresh. Cheap
// (one issue + one service/issue key); the count tables are reconciled by the full
// refresh. It takes refreshMu so it cannot run while a full refresh is mid build-and-
// swap: otherwise a swap of a snapshot taken before the rescore committed would revert
// this write until the next full refresh.
func (mve *MvEngine) TargetedVulnerabilityRefresh(serviceID, issueID int64) {
	db, err := NewDb(mve.cfg)
	if err != nil {
		logrus.WithError(err).Error("targeted vulnerability MV refresh: could not create db")
		return
	}

	defer func() {
		_ = db.Close()
	}()

	mve.refreshMu.Lock()
	err = RefreshVulnerabilityForRemediation(context.Background(), db, serviceID, issueID)
	mve.refreshMu.Unlock()

	if err != nil {
		logrus.WithError(err).Warn("targeted vulnerability MV refresh failed; full refresh will reconcile")
		return
	}

	if cb := mve.OnRefreshComplete; cb != nil {
		cb()
	}
}

////////// Internals

type mveCtx struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []string
}

func (mc *mveCtx) appendErrorMessage(msg string) {
	mc.mu.Lock()
	mc.errs = append(mc.errs, msg)
	mc.mu.Unlock()
}

func (mc *mveCtx) hasError() bool {
	return len(mc.errs) > 0
}

func (mc *mveCtx) getError() error {
	return fmt.Errorf("error when execute joined callers: [%s]", strings.Join(mc.errs, "; "))
}

func (mc *mveCtx) Wait() error {
	mc.wg.Wait()

	if mc.hasError() {
		return mc.getError()
	}

	return nil
}

func runInBackground(db Db, procs [][]MVProcedure) *mveCtx {
	mc := &mveCtx{}

	for i, pl := range procs {
		mc.wg.Go(func() {
			for j, p := range pl {
				if err := TxCall(p, context.Background(), db); err != nil {
					mc.appendErrorMessage(fmt.Sprintf("(procIdx: %d:%d): %v", i, j, err))
					break
				}
			}
		})
	}

	runCleanupRoutineInBackground(db, mc)

	return mc
}

func TxCall(p MVProcedure, ctx context.Context, db Db) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	err = p(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func runCleanupRoutineInBackground(db Db, mc *mveCtx) {
	go func() {
		if err := mc.Wait(); err != nil {
			logrus.WithError(err).Error("MVE background execution failed")
		}

		_ = db.Close()
	}()
}
