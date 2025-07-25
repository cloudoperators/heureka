// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type CacheConfig struct {
	KeyHash                  KeyHashType
	MonitorInterval          time.Duration
	MaxDbConcurrentRefreshes int
	ThrottleInterval         time.Duration
	ThrottlePerInterval      int
}

type Stat struct {
	Hit  int64
	Miss int64
}

type CacheBase struct {
	stat                       Stat
	keyHash                    KeyHashType
	statMu                     sync.RWMutex
	monitorMu                  sync.Mutex
	monitorOnce                sync.Once
	concurrentRefreshSemaphore chan struct{}
	concurrentRefreshUnlimited bool
	refreshLimiter             *rate.Limiter
	ctx                        context.Context
	wg                         *sync.WaitGroup
}

func NewCacheBase(ctx context.Context, wg *sync.WaitGroup, config CacheConfig) *CacheBase {
	cb := CacheBase{keyHash: config.KeyHash, ctx: ctx, wg: wg}
	cb.initConcurrentRefreshLimit(config)
	cb.initRateRefreshLimit(config)
	return &cb
}

func (cb *CacheBase) initConcurrentRefreshLimit(config CacheConfig) {
	if config.MaxDbConcurrentRefreshes > 0 {
		cb.concurrentRefreshSemaphore = make(chan struct{}, config.MaxDbConcurrentRefreshes)
	} else if config.MaxDbConcurrentRefreshes < 0 {
		cb.concurrentRefreshUnlimited = true
	}
}

func (cb *CacheBase) initRateRefreshLimit(config CacheConfig) {
	if config.ThrottleInterval > 0 {
		cb.refreshLimiter = rate.NewLimiter(rate.Every(config.ThrottleInterval), config.ThrottlePerInterval)
	}
}

func (cb *CacheBase) startMonitorIfNeeded(interval time.Duration) {
	if interval <= 0 {
		return
	}
	cb.monitorOnce.Do(func() {
		l := logrus.New()
		cb.monitorMu.Lock()
		defer cb.monitorMu.Unlock()

		cb.wg.Add(1)
		go func() {
			defer cb.wg.Done()
			l.Info("Monitoring started with interval: ", interval)
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-cb.ctx.Done():
					l.Info("Monitoring stopped")
					return
				case <-ticker.C:
					l.Info(StatStr(cb.GetStat()))
				}
			}
		}()
	})
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

func (cb CacheBase) EncodeKey(key string) string {
	if cb.keyHash == KEY_HASH_SHA256 {
		return encodeSHA256(key)
	} else if cb.keyHash == KEY_HASH_SHA512 {
		return encodeSHA512(key)
	} else if cb.keyHash == KEY_HASH_HEX {
		return encodeHex(key)
	} else if cb.keyHash == KEY_HASH_NONE {
		return key
	}
	return encodeBase64(key)
}

func (cb CacheBase) CacheKey(fnname string, fn interface{}, args ...interface{}) (string, error) {
	key, err := cacheKeyJson(fnname, fn, args...)
	if err != nil {
		return "", fmt.Errorf("Cache: could not create json cache key.")
	}
	return cb.EncodeKey(key), nil
}

func (cb *CacheBase) launchRefreshWithThrottling(fn func()) {
	if cb.refreshLimiter == nil || cb.refreshLimiter.Allow() {
		cb.wg.Add(1)
		go func() {
			defer cb.wg.Done()
			fn()
		}()
	}
}

func (cb *CacheBase) LaunchRefresh(fn func()) {
	if cb.concurrentRefreshUnlimited {
		cb.launchRefreshWithThrottling(fn)
	} else if cb.concurrentRefreshSemaphore != nil {
		select {
		case cb.concurrentRefreshSemaphore <- struct{}{}:
			cb.launchRefreshWithThrottling(func() {
				fn()
				<-cb.concurrentRefreshSemaphore
			})
		default:
			// Optional: log or track skipped refresh due to throttling
		}
	}
}

func cacheKeyJson(fnname string, fn interface{}, args ...interface{}) (string, error) {
	keyParts := make([]interface{}, 0, len(args)+1)
	keyParts = append(keyParts, fnname)

	for i, arg := range args {
		if !isJSONSerializable(arg) {
			return "", fmt.Errorf("argument %d is not JSON serializable: %T", i, arg)
		}
		keyParts = append(keyParts, arg)
	}

	// Encode the full key as JSON array
	jsonKey, err := json.Marshal(keyParts)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cache key: %w", err)
	}

	return string(jsonKey), nil
}

func isJSONSerializable(val interface{}) bool {
	_, err := json.Marshal(val)
	return err == nil
}

func StatStr(stat Stat) string {
	var hmr float32
	total := stat.Hit + stat.Miss
	if total > 0 {
		hmr = float32(stat.Hit) / float32(total)
	}
	return fmt.Sprintf("hit: %d, miss: %d, h/(h+m): %f", stat.Hit, stat.Miss, hmr)
}
