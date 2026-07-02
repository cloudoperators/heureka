// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/go-co-op/gocron"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var MVProcedures = []string{
	"refresh_mvServiceIssueCounts_proc",
	"refresh_mvSingleComponentByServiceVulnerabilityCounts_proc",
	"refresh_mvAllComponentsByServiceVulnerabilityCounts_proc",
	"refresh_mvCountIssueRatingsUniqueService_proc",
	"refresh_mvCountIssueRatingsService_proc",
	"refresh_mvCountIssueRatingsServiceWithoutSupportGroup_proc",
	"refresh_mvCountIssueRatingsSupportGroup_proc",
	"refresh_mvCountIssueRatingsComponentVersion_proc",
	"refresh_mvCountIssueRatingsServiceId_proc",
	"refresh_mvCountIssueRatingsOther_proc",
	"refresh_mvVulnerabilityList_proc",
	"refresh_mvVulnerabilityService_proc",
	"refresh_mvComponentService_proc",
}

func TriggerMVE(cfg util.Config) error {
	db, err := GetSqlxConnection(cfg)
	if err != nil {
		return fmt.Errorf("error while Creating Db: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logrus.Warnf("failed to close DB connection: %s", err)
		}
	}()

	if err := checkProceduresExist(db, MVProcedures); err != nil {
		return err
	}

	return runInBackground(db, MVProcedures).Wait()
}

func StartMVEScheduler(cfg util.Config) {
	mve := getMVE()

	periodMinutes := cfg.DBMvCalcPeriodMinutes
	if periodMinutes <= 0 {
		periodMinutes = 200
	}

	logrus.Debugf("MVE scheduling period set to %d minutes", periodMinutes)

	_, err := mve.scheduler.Every(periodMinutes).Minutes().SingletonMode().Do(func() {
		err := TriggerMVE(cfg)
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

func StopMVE() {
	mve := getMVE()
	mve.Shutdown()
}

func WaitMVEForFirstRun() {
	mve := getMVE()
	<-mve.firstRunDone
}

////////// Internals

var mvEngine *MvEngine

type MvEngine struct {
	scheduler    *gocron.Scheduler
	firstRunDone chan struct{}
	once         sync.Once
}

func NewMvEngine() *MvEngine {
	return &MvEngine{
		scheduler:    gocron.NewScheduler(time.UTC),
		firstRunDone: make(chan struct{}),
	}
}

func getMVE() *MvEngine {
	if mvEngine == nil {
		mvEngine = NewMvEngine()
	}

	return mvEngine
}

func (mve *MvEngine) Shutdown() {
	mve.scheduler.Stop()
}

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

func checkProceduresExist(db *sqlx.DB, procs []string) error {
	exists, err := proceduresExist(db, procs)
	if err != nil {
		return fmt.Errorf("could not check procedures exist: %w", err)
	} else if !exists {
		return fmt.Errorf("some procedures [%s] do not exist", strings.Join(procs, ", "))
	}

	return nil
}

func proceduresExist(db *sqlx.DB, procedures []string) (bool, error) {
	if len(procedures) == 0 {
		return true, nil
	}

	placeholders := make([]string, len(procedures))
	args := make([]any, len(procedures))

	for i, p := range procedures {
		placeholders[i] = "?"
		args[i] = p
	}

	query := fmt.Sprintf(`
        SELECT COUNT(*)
        FROM information_schema.routines
        WHERE routine_schema = DATABASE()
          AND routine_type = 'PROCEDURE'
          AND routine_name IN (%s);
    `, strings.Join(placeholders, ","))

	var count int

	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("could not check if procedures exist: %w", err)
	}

	if count == len(procedures) {
		return true, nil
	}

	return false, nil
}

func runInBackground(db *sqlx.DB, procs []string) *mveCtx {
	mc := &mveCtx{}

	for _, p := range procs {
		mc.wg.Go(func() {
			if _, err := db.Exec(fmt.Sprintf("CALL %s();", p)); err != nil {
				mc.appendErrorMessage(fmt.Sprintf("%s: %v", p, err))
			}
		})
	}

	runCleanupRoutineInBackground(db, mc)

	return mc
}

func runCleanupRoutineInBackground(db *sqlx.DB, mc *mveCtx) {
	go func() {
		if err := mc.Wait(); err != nil {
			logrus.WithError(err).Error(err)
		}

		_ = db.Close()
	}()
}
