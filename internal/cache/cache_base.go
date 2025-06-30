package cache

import (
	"fmt"
	"sync"
	"time"
)

type CacheConfig struct {
	Ttl     time.Duration
	KeyHash KeyHashType
}

type Stat struct {
	Hit  int64
	Miss int64
}

type CacheBase struct {
	stat    Stat
	keyHash KeyHashType
	ttl     time.Duration
	statMu  sync.RWMutex
}

func NewCacheBase(config CacheConfig) *CacheBase {
	return &CacheBase{ttl: config.Ttl, keyHash: config.KeyHash}
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
