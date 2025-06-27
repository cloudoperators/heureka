// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

        //cache:         cache.NewCache(cache.InMemoryCacheConfig{Ttl: 1 * time.Minute}),
        //cache:         cache.NewCache(cache.RedisCacheConfig{Url: "redis:6378", Ttl: 1 * time.Minute}), //wrong port
        //cache:         cache.NewCache(cache.RedisCacheConfig{Url: "localhost:6379", Ttl: 1 * time.Minute}),
        //cache:         cache.NewCache(cache.RedisCacheConfig{Url: "localhost:6379", Ttl: 1 * time.Nanosecond}),
        //cache:         cache.NewCache(cache.RedisCacheConfig{Url: "redis:6379"}),
        //cache:         cache.NewCache(cache.Config{Ttl: 1*time.Minute, KeyHash: cache.KEY_HASH_SHA256}), //TODO: read from envconfig???

type cacheTest struct {
	seeder *test.DatabaseSeeder
	server *server.Server
	cfg util.Config
	db *mariadb.SqlDatabase
	seedCollection *test.SeedCollection
}

func NewCacheTest(redisUrl string, ttlMSec int64) *cacheTest {
	var ct cacheTest
	ct.db = dbm.NewTestSchema()

	var err error
	ct.seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")

	ct.cfg = dbm.DbConfig()
	ct.cfg.Port = util2.GetRandomFreePort()

	ct.cfg.CacheTtlMSec = ttlMSec
	ct.cfg.CacheRedisUrl = redisUrl

	ct.server = server.NewServer(ct.cfg)
	ct.server.NonBlockingStart()

	ct.seedCollection = ct.seeder.SeedDbWithNFakeData(1)
	return &ct
}

func (ct *cacheTest)teardown() {
	ct.server.BlockingStop()
	dbm.TestTearDown(ct.db)
}

func (ct *cacheTest)queryTestResource() {
	existingServiceCcrns := lo.Map(ct.seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, index int) string {
		return s.CCRN.String
	})
	queryComponentInstanceFilterAndExpectVal(
		ct.cfg.Port,
		"../api/graphql/graph/queryCollection/componentInstanceFilter/serviceCcrn.graphqls",
		existingServiceCcrns,
		func(cifv model.ComponentInstanceFilterValue) []*string {
			return cifv.ServiceCcrn.Values
	})
}

var _ = Describe("Using Redis cache", Label("e2e", "RedisCache"), func() { //CHECK MISS
	Describe("Check miss", func() {
		var ct *cacheTest
		AfterEach(func() {
			ct.teardown()
		})
		Context("Redis cache is configured", func() {
			BeforeEach(func() {
				ct = NewCacheTest("", 0) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//ct = NewCacheTest("", 24 * 60 * 60 * 1000) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//ct = NewCacheTest("localhost:6379", 24 * 60 * 60 * 1000) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//configureRedisCache()
			})
			When("Test resource is queried for the first time", func() {
				BeforeEach(func() {
					ct.queryTestResource()
				})
				It("Miss counter should be equal 1 and Hit counter should be equal to 0", func() {
					//expectMissCounter(1)
					//expectHitCounter(0)
				})
			})
		})
	})
	Describe("Check hit", func() {
		var ct *cacheTest
		AfterEach(func() {
			ct.teardown()
		})
		Context("Redis cache is configured and cache contains test resource", func() {
			BeforeEach(func() {
				ct = NewCacheTest("", 0) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//ct = NewCacheTest("", 24 * 60 * 60 * 1000) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//ct = NewCacheTest("localhost:6379", 24 * 60 * 60 * 1000) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//configureRedisCache()
				ct.queryTestResource()
			})
			When("Cached test resource is queried", func() {
				BeforeEach(func() {
					ct.queryTestResource()
				})
				It("Hit counter should be equal 1 and Miss counter should be equal to 0", func() {
					//expectHitCounter(1)
					//expectMissCounter(0)
				})
			})
		})
	})
	Describe("Check expired", func() {
		var ct *cacheTest
		AfterEach(func() {
			ct.teardown()
		})
		Context("Redis cache is configured with short TTL duration and test resource is queried for the first time", func() {
			BeforeEach(func() {
				ct = NewCacheTest("", 0) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//ct = NewCacheTest("", 1) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//ct = NewCacheTest("localhost:6379", 1) //TODO: port has to be configured using 'util2.GetRandomFreePort()'
				//configureRedisCacheWithShortTtl()
				ct.queryTestResource()
			})
			When("TTL duration elapsed and test resource is queried", func() {
				BeforeEach(func() {
					//sleepTtlDuration()
					ct.queryTestResource()
				})
				It("Expired counter should be equal 1 and Miss counter should be equal to 2 and Hit counter should be equal 0", func() {
					//expectExpiredCounter(1)
					//expectMissCounter(2)
					//expectHitCounter(0)
				})
			})
		})
	})
})

//var _ = Describe("Using In memory cache", Label("e2e", "InMemoryCache"), func() {
//})

//var _ = Describe("Using No cache", Label("e2e", "NoCache"), func() {
//})


// CHECK MISS
// Given Redis cache is configured
//  When Test resource is queried for the first time 
//  Then Miss counter should be equal 1
//   And Hit counter should be equal 0

// CHECK HIT
// Given Redis cache is configured
//   And Cache contain test resource
//  When Test resource is queried
//  Then Hit counter should be equal 1
//   And Miss counter should be equal 0

// CHECK EXPIRED
// Given Redis cache is configured with short TTL duration
//   And Test resource is queried for the first time
//  When TTL duration elapsed
//   And Test resource is queried
//  Then Expired counter should be equal 1
//   And Miss counter should be equal 2
//   And Hit counter should be equal 0



// cache config:
// CacheTtlMSec
// CacheRedisUrl

// if CacheTtlMSec is equal 0 -> NoCache
//   if CacheRedisUrl is set -> RedisCache
//   else -> InMemoryCache
