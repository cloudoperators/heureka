package cache

import (
	"time"
)

func NewNoCache() *NoCache {
	return &NoCache{}
}

type NoCache struct {
}

func (nc NoCache) CacheKey(fnname string, fn interface{}, args ...interface{}) (string, error) {
	return "", nil
}

func (nc NoCache) Get(key string) (string, bool, error) {
	return "", false, nil
}

func (nc NoCache) Set(key string, value string, ttl time.Duration) error {
	return nil
}

func (nc NoCache) Invalidate(key string) error {
	return nil
}

func (nc NoCache) IncHit() {
}

func (nc NoCache) IncMiss() {
}

func (nc NoCache) GetStat() Stat {
	return Stat{}
}

func (nc NoCache) LaunchRefresh(func()) {
}
