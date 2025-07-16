// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/app/service"
	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

const (
	defaultTtl                      = 24 * time.Hour
	ttl1Millisecond                 = time.Millisecond
	testResourceCount               = 1
	addResourceCount                = 1
	noConcurrentLimit               = -1
	noThrottleIntervalMSec          = 0
	noThrottlePerInterval           = 1
	noCacheTtl                      = 0
	valkeyUrl                       = "localhost:6379"
	shortTtlTimeToWait              = 10 * time.Millisecond
	backgroundUpdateTimeToWait      = 100 * time.Millisecond
	backgroundUpdateStartTimeToWait = 50 * time.Millisecond
)

type cacheTest struct {
	seeder              *test.DatabaseSeeder
	server              *server.Server
	cfg                 util.Config
	db                  *mariadb.SqlDatabase
	seedCollection      *test.SeedCollection
	lastResource        model.ComponentInstanceFilterValue
	addedSeedCollection *test.SeedCollection
	dbProxy             *e2e_common.PausableProxy
}

type cacheConfig struct {
	valkeyUrl string
}

var valkeyCacheConfig = cacheConfig{valkeyUrl: valkeyUrl}
var inMemoryCacheConfig = cacheConfig{}

func newCacheTest(valkeyUrl string, ttl time.Duration, cacheEnable bool, maxDbConcurrentRefreshes int, throttleIntervalMSec int64, throttlePerInterval int) *cacheTest {
	var ct cacheTest
	ct.db = dbm.NewTestSchema()

	var err error
	ct.seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")

	ct.cfg = dbm.DbConfig()
	ct.cfg.Port = util2.GetRandomFreePort()
	ct.cfg.CacheEnable = cacheEnable
	ct.cfg.CacheValkeyUrl = valkeyUrl
	ct.cfg.CacheMaxDbConcurrentRefreshes = maxDbConcurrentRefreshes
	ct.cfg.CacheThrottleIntervalMSec = throttleIntervalMSec
	ct.cfg.CacheThrottlePerInterval = throttlePerInterval

	ct.server = server.NewServer(ct.cfg)
	ct.server.NonBlockingStart()

	ct.seedCollection = ct.seeder.SeedDbWithNFakeData(testResourceCount)

	service.CacheTtlGetServiceCcrns = ttl
	return &ct
}

func newCacheTestWithDbProxy(valkeyUrl string, ttl time.Duration, cacheEnable bool, maxDbConcurrentRefreshes int, throttleIntervalMSec int64, throttlePerInterval int) *cacheTest {
	var ct cacheTest
	ct.db = dbm.NewTestSchema()

	var err error
	ct.seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")

	ct.cfg = dbm.DbConfig()

	dbPort := ct.cfg.DBPort
	dbProxyPort := util2.GetRandomFreePort()
	ct.dbProxy = e2e_common.NewPausableProxy(fmt.Sprintf("localhost:%s", dbProxyPort), fmt.Sprintf("localhost:%s", dbPort))
	err = ct.dbProxy.Start()
	Expect(err).To(BeNil(), "Could not start DB proxy")
	ct.cfg.DBPort = dbProxyPort
	ct.cfg.DBMaxIdleConnections = 0

	ct.cfg.Port = util2.GetRandomFreePort()
	ct.cfg.CacheEnable = cacheEnable
	ct.cfg.CacheValkeyUrl = valkeyUrl
	ct.cfg.CacheMaxDbConcurrentRefreshes = maxDbConcurrentRefreshes
	ct.cfg.CacheThrottleIntervalMSec = throttleIntervalMSec
	ct.cfg.CacheThrottlePerInterval = throttlePerInterval

	ct.server = server.NewServer(ct.cfg)
	ct.server.NonBlockingStart()

	ct.seedCollection = ct.seeder.SeedDbWithNFakeData(testResourceCount)

	service.CacheTtlGetServiceCcrns = ttl
	return &ct
}

func newNoCacheTest() *cacheTest {
	return newCacheTest("", noCacheTtl, false, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval)
}

func (ct *cacheTest) teardown() {
	ct.server.BlockingStop()
	dbm.TestTearDown(ct.db)
	if ct.dbProxy != nil {
		ct.dbProxy.Stop()
	}
}

func (ct *cacheTest) testResourceIsQueried() {
	ct.queryResource()
	ct.expectTestResource()
}

func (ct *cacheTest) expectTestResource() {
	existingServiceCcrns := lo.Map(ct.seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, index int) string {
		return s.CCRN.String
	})
	expectResponseVal(
		ct.lastResource,
		existingServiceCcrns,
		func(cifv model.ComponentInstanceFilterValue) []*string {
			return cifv.ServiceCcrn.Values
		})
}

func (ct *cacheTest) expectTestResourceAndAddedResource() {
	Expect(len(ct.addedSeedCollection.ServiceRows)).To(Equal(addResourceCount))
	existingServiceCcrns := lo.Map(append(ct.seedCollection.ServiceRows, ct.addedSeedCollection.ServiceRows...), func(s mariadb.BaseServiceRow, index int) string {
		return s.CCRN.String
	})
	expectResponseVal(
		ct.lastResource,
		existingServiceCcrns,
		func(cifv model.ComponentInstanceFilterValue) []*string {
			return cifv.ServiceCcrn.Values
		})
}

func (ct *cacheTest) addDbResource() {
	ct.addedSeedCollection = ct.seeder.SeedDbWithNFakeData(addResourceCount)
}

func (ct *cacheTest) queryResource() {
	ct.lastResource = queryComponentInstanceFilter(
		ct.cfg.Port,
		"../api/graphql/graph/queryCollection/componentInstanceFilter/serviceCcrn.graphqls",
	)
}

func (ct *cacheTest) expectMissHitCounter(expectedMiss, expectedHit int64) {
	stat := ct.server.App().GetCache().GetStat()
	Expect(stat.Miss).To(Equal(expectedMiss))
	Expect(stat.Hit).To(Equal(expectedHit))
}

func missTest(config cacheConfig) {
	// GIVEN Cache is configured
	ct := newCacheTest(config.valkeyUrl, defaultTtl, true, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval)
	defer ct.teardown()

	// WHEN Test resource is queried for the first time
	ct.testResourceIsQueried()

	// THEN Miss = 1 and Hit = 0
	ct.expectMissHitCounter(1, 0)
}

func hitTest(config cacheConfig) {
	// GIVEN Cache is configured
	ct := newCacheTest(config.valkeyUrl, defaultTtl, true, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval)
	defer ct.teardown()

	// AND Cache contains test resource
	ct.testResourceIsQueried()

	// WHEN Cached test resource is queried
	ct.testResourceIsQueried()

	// THEN Miss = 1 and Hit = 1
	ct.expectMissHitCounter(1, 1)
}

func expiredTest(config cacheConfig) {
	// GIVEN Cache is configured with short TTL
	ct := newCacheTest(config.valkeyUrl, ttl1Millisecond, true, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval)
	defer ct.teardown()

	// AND Test resource is queried for the first time
	ct.testResourceIsQueried()

	// WHEN TTL duration has passed
	time.Sleep(shortTtlTimeToWait)

	// AND Test resource is queried
	ct.testResourceIsQueried()

	// THEN Miss = 2 and Hit = 0
	ct.expectMissHitCounter(2, 0)
}

func backgroundUpdateTest(config cacheConfig) {
	// GIVEN Cache is configured
	ct := newCacheTest(config.valkeyUrl, defaultTtl, true, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval)
	defer ct.teardown()

	// AND Test resource is queried for the first time
	ct.testResourceIsQueried()

	// AND Extra DB resource is added
	ct.addDbResource()

	// AND Test resource is queried
	ct.testResourceIsQueried()

	// AND Miss = 1 and Hit = 1
	ct.expectMissHitCounter(1, 1)

	// WHEN Wait for a moment for background cache update
	time.Sleep(backgroundUpdateTimeToWait)

	// AND Resource is queried
	ct.queryResource()

	// THEN Resource contains test resource and added resource
	ct.expectTestResourceAndAddedResource()

	// AND Miss = 1 and Hit = 2
	ct.expectMissHitCounter(1, 2)
}

func backgroundUpdateRateBasedLimitTest(config cacheConfig) {

	// -- Skip background update above the limit

	// GIVEN Cache is configured with background update limit set to 2 per 1 second
	ct := newCacheTest(config.valkeyUrl, defaultTtl, true, noConcurrentLimit, 1000, 2)
	defer ct.teardown()

	// AND Three times resource is queried to update cache with test resource on miss and exhaust the limit of background updates
	ct.queryResource()
	ct.queryResource()
	ct.queryResource()

	// AND Wait for a moment for background cache update
	time.Sleep(backgroundUpdateTimeToWait)

	// WHEN Extra DB resource is added
	ct.addDbResource()

	// AND Resource is queried
	ct.queryResource()

	// AND Wait for a moment to be sure background refresh is skipped
	time.Sleep(backgroundUpdateTimeToWait)

	// AND Resource is queried
	ct.queryResource()

	// THEN Resource contains test resource
	ct.expectTestResource()

	// AND Miss = 1 and Hit = 4
	ct.expectMissHitCounter(1, 4)

	// -- Allow background update after rate limit interval

	// GIVEN Wait for rate limit interval to be able to update cache in background again
	time.Sleep(1 * time.Second)

	// AND Test resource is queried
	ct.testResourceIsQueried()

	// AND Miss = 1 and Hit = 5
	ct.expectMissHitCounter(1, 5)

	// WHEN Wait for a moment for background cache update
	time.Sleep(backgroundUpdateTimeToWait)

	// AND Resource is queried
	ct.queryResource()

	// THEN Resource contains test resource and added resource
	ct.expectTestResourceAndAddedResource()

	// AND Miss = 1 and Hit = 6
	ct.expectMissHitCounter(1, 6)
}

func backgroundUpdateConcurrentLimitTest(config cacheConfig) {

	// -- Skip background update above the limit

	// GIVEN Cache is configured with DB proxy and with background update concurrent limit set to 2
	ct := newCacheTestWithDbProxy(config.valkeyUrl, defaultTtl, true, 2, noThrottleIntervalMSec, noThrottlePerInterval)
	defer ct.teardown()

	// AND Resource is queried to update cache with test resource on miss
	ct.queryResource()
	ct.expectMissHitCounter(1, 0)

	// AND DB access is paused for new connections
	ct.dbProxy.PauseConnections()

	// AND Two times resource is queried to exhaust the limit of concurrent background updates
	ct.queryResource()
	ct.queryResource()

	// AND Wait for a moment to start background updates
	time.Sleep(backgroundUpdateStartTimeToWait)

	// AND DB access is allowed for new connections
	ct.dbProxy.ResumeConnections()

	// WHEN Extra DB resource is added
	ct.addDbResource()

	// AND Resource is queried
	ct.queryResource()

	// AND Wait for a moment to be sure background refresh is skipped
	time.Sleep(backgroundUpdateTimeToWait)

	// AND Resource is queried
	ct.queryResource()

	// THEN Resource contains test resource
	ct.expectTestResource()

	// AND Miss = 1 and Hit = 4
	ct.expectMissHitCounter(1, 4)

	// -- Allow background update when below concurrent limit

	// GIVEN close all hanged DB connections
	ct.dbProxy.CloseHeldConnections()

	// AND Test resource is queried
	ct.testResourceIsQueried()

	// AND Miss = 1 and Hit = 5
	ct.expectMissHitCounter(1, 5)

	// WHEN Wait for a moment for background cache update
	time.Sleep(backgroundUpdateTimeToWait)

	// AND Resource is queried
	ct.queryResource()

	// THEN Resource contains test resource and added resource
	ct.expectTestResourceAndAddedResource()

	// AND Miss = 1 and Hit = 6
	ct.expectMissHitCounter(1, 6)
}

var _ = Describe("Using Valkey cache", Label("e2e", "ValkeyCache"), Label("e2e", "Cache"), func() {

	Describe("Check miss", func() {
		It("Should increase miss counter when resource is queried for the first time", func() {
			missTest(valkeyCacheConfig)
		})
	})

	Describe("Check hit", func() {
		It("Should increase hit counter when the same resource is queried for the second time", func() {
			hitTest(valkeyCacheConfig)
		})
	})

	Describe("Check expired", func() {
		It("Should increase miss counter when the same resource is queried after TTL duration", func() {
			expiredTest(valkeyCacheConfig)
		})
	})

	Describe("Check background update", func() {
		It("Should update resource in background on hit so next hit will show new data", func() {
			backgroundUpdateTest(valkeyCacheConfig)
		})
	})

	Describe("Check background update rate-based limit", func() {
		It("Should skip background updates above the rate limit AND execute new background updates after interval", func() {
			backgroundUpdateRateBasedLimitTest(valkeyCacheConfig)
		})
	})

	Describe("Check background update concurrent-based limit", func() {
		It("Should skip background updates above the concurrent limit AND execute new background updates when first background updates are over", func() {
			backgroundUpdateConcurrentLimitTest(valkeyCacheConfig)
		})
	})
})

var _ = Describe("Using In memory cache", Label("e2e", "InMemoryCache"), Label("e2e", "Cache"), func() {

	Describe("Check miss", func() {
		It("Should increase miss counter when resource is queried for the first time", func() {
			missTest(inMemoryCacheConfig)
		})
	})

	Describe("Check hit", func() {
		It("Should increase hit counter when the same resource is queried for the second time", func() {
			hitTest(inMemoryCacheConfig)
		})
	})

	Describe("Check expired", func() {
		It("Should increase miss counter when the same resource is queried after TTL duration", func() {
			expiredTest(inMemoryCacheConfig)
		})
	})

	Describe("Check background update", func() {
		It("Should update resource in background on hit so next hit will show new data", func() {
			backgroundUpdateTest(inMemoryCacheConfig)
		})
	})

	Describe("Check background update rate-based limit", func() {
		It("Should skip background updates above the rate limit AND execute new background updates after interval", func() {
			backgroundUpdateRateBasedLimitTest(inMemoryCacheConfig)
		})
	})

	Describe("Check background update concurrent-based limit", func() {
		It("Should skip background updates above the concurrent limit AND execute new background updates when first background updates are over", func() {
			backgroundUpdateConcurrentLimitTest(inMemoryCacheConfig)
		})
	})
})

var _ = Describe("Using No cache", Label("e2e", "NoCache"), Label("e2e", "Cache"), func() {

	Describe("Check miss", func() {
		It("Should keep miss counter 0 when resource is queried for the first time", func() {

			// GIVEN No cache is configured
			ct := newNoCacheTest()
			defer ct.teardown()

			// WHEN Test resource is queried for the first time
			ct.testResourceIsQueried()

			// THEN Miss = 0 and Hit = 0
			ct.expectMissHitCounter(0, 0)
		})
	})

	Describe("Check hit", func() {
		It("Should keep hit counter 0 when the same resource is queried for the second time", func() {

			// GIVEN No cache is configured
			ct := newNoCacheTest()
			defer ct.teardown()

			// AND cache contains test resource
			ct.testResourceIsQueried()

			// WHEN Cached test resource is queried
			ct.testResourceIsQueried()

			// THEN Miss = 0 and Hit = 0
			ct.expectMissHitCounter(0, 0)
		})
	})
})
