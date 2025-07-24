// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valkey-io/valkey-go"
)

type ValkeyCache struct {
	CacheBase
	client valkey.Client
}

type ValkeyCacheConfig struct {
	CacheConfig
	Url      string
	Password string
	Db       int
}

func NewValkeyCache(ctx context.Context, wg *sync.WaitGroup, config ValkeyCacheConfig) *ValkeyCache {
	cacheBase := NewCacheBase(ctx, wg, config.CacheConfig)
	valkeyClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{config.Url},
		Password:    config.Password})
	if err != nil {
		log := logrus.New()
		log.WithFields(logrus.Fields{
			"component": "cache",
			"error":     err,
		}).Fatal("Failed to initialize Valkey cache")
		return nil
	}
	valkeyCache := &ValkeyCache{
		CacheBase: *cacheBase,
		client:    valkeyClient,
	}

	valkeyCache.startMonitorIfNeeded(config.MonitorInterval)

	_ = valkeyCache.invalidateAll()

	return valkeyCache
}

func (vc *ValkeyCache) Get(key string) (string, bool, error) {
	val, err := vc.client.Do(vc.ctx, vc.client.B().Get().Key(key).Build()).ToString()
	if err == valkey.Nil {
		return "", false, nil // miss
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// ttl = 0 <- infinite
func (vc *ValkeyCache) Set(key string, value string, ttl time.Duration) error {
	return vc.client.Do(vc.ctx, vc.client.B().Set().Key(key).Value(value).Px(ttl).Build()).Error()
}

func (vc *ValkeyCache) Invalidate(key string) error {
	return vc.client.Do(vc.ctx, vc.client.B().Del().Key(key).Build()).Error()
}

func (vc *ValkeyCache) invalidateAll() error {
	return vc.client.Do(vc.ctx, vc.client.B().Flushall().Build()).Error()
}
