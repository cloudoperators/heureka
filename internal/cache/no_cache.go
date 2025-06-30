package cache

import (
	"context"
)

func NewNoCache() *NoCache {
	return &NoCache{}
}

type NoCache struct {
}

func (nc NoCache) cacheKey(fnname string, fn interface{}, args ...interface{}) (string, error) {
	return "", nil
}

func (nc NoCache) Get(ctx context.Context, key string) (string, bool, error) {
	return "", false, nil
}

func (nc NoCache) Set(ctx context.Context, key string, value string) error {
	return nil
}

func (nc NoCache) Invalidate(ctx context.Context, key string) error {
	return nil
}

func (nc *NoCache) IncHit() {
}

func (nc *NoCache) IncMiss() {
}

func (nc NoCache) GetStat() Stat {
	return Stat{}
}
