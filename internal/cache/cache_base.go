package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CacheConfig struct {
	Ttl             time.Duration
	KeyHash         KeyHashType
	MonitorInterval time.Duration
}

type Stat struct {
	Hit  int64
	Miss int64
}

type CacheBase struct {
	stat                  Stat
	keyHash               KeyHashType
	ttl                   time.Duration
	statMu                sync.RWMutex
	monitorMu             sync.Mutex
	monitorCancelFunction context.CancelFunc
	monitorCtx            context.Context
}

func NewCacheBase(config CacheConfig) *CacheBase {
	return &CacheBase{ttl: config.Ttl, keyHash: config.KeyHash}
}

func (cb *CacheBase) startMonitorIfNeeded(interval time.Duration) {
	if interval <= 0 {
		return
	}
	l := logrus.New()
	cb.monitorMu.Lock()
	defer cb.monitorMu.Unlock()

	if cb.monitorCancelFunction != nil {
		l.Error("Monitoring already started.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	cb.monitorCtx = ctx
	cb.monitorCancelFunction = cancel

	go func() {
		l.Info("Monitoring started with interval: ", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				l.Info("Monitoring stopped")
				return
			case <-ticker.C:
				l.Info(StatStr(cb.GetStat()))
			}
		}
	}()
}

func (cb *CacheBase) Stop() {
	cb.monitorMu.Lock()
	defer cb.monitorMu.Unlock()

	if cb.monitorCancelFunction != nil {
		cb.monitorCancelFunction()
	}
}

func (cb *CacheBase) IncHit() {
	cb.statMu.Lock()
	defer cb.statMu.Unlock()
	cb.stat.Hit = cb.stat.Hit + 1
}

func (cb *CacheBase) IncMiss() {
	cb.statMu.Lock()
	defer cb.statMu.Unlock()
	cb.stat.Miss = cb.stat.Miss + 1
}

func (cb CacheBase) GetStat() Stat {
	cb.statMu.RLock()
	defer cb.statMu.RUnlock()
	return cb.stat
}

func StatStr(stat Stat) string {
	var hmr float32
	total := stat.Hit + stat.Miss
	if total > 0 {
		hmr = float32(stat.Hit) / float32(total)
	}
	return fmt.Sprintf("hit: %d, miss: %d, h/(h+m): %f", stat.Hit, stat.Miss, hmr)
}
