// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"time"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"

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
	noProxy                         = false
	withProxy                       = true
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

var (
	valkeyCacheTestConfig   = cacheConfig{valkeyUrl: valkeyUrl}
	inMemoryCacheTestConfig = cacheConfig{}
)

func newCacheTest(valkeyUrl string, ttl time.Duration, maxDbConcurrentRefreshes int, throttleIntervalMSec int64, throttlePerInterval int, proxy bool) *cacheTest {
	var ct cacheTest
	ct.db = dbm.NewTestSchemaWithoutMigration()

	var err error
	ct.seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil(), "Database Seeder Setup should work")

	ct.cfg = dbm.DbConfig()

	if proxy {
		ct.startDbProxy()
	}

	ct.cfg.Port = e2e_common.GetRandomFreePort()
	ct.cfg.CacheEnable = true
	ct.cfg.CacheValkeyUrl = valkeyUrl
	ct.cfg.CacheMaxDbConcurrentRefreshes = maxDbConcurrentRefreshes
	ct.cfg.CacheThrottleIntervalMSec = throttleIntervalMSec
	ct.cfg.CacheThrottlePerInterval = throttlePerInterval

	ct.server = e2e_common.NewRunningServer(ct.cfg)

	ct.seedCollection = ct.seeder.SeedDbWithNFakeData(testResourceCount)

	service.CacheTtlGetServiceAttrs = ttl
	return &ct
}

func (ct *cacheTest) startDbProxy() {
	dbPort := ct.cfg.DBPort
	dbProxyPort := e2e_common.GetRandomFreePort()
	ct.dbProxy = e2e_common.NewPausableProxy(fmt.Sprintf("localhost:%s", dbProxyPort), fmt.Sprintf("localhost:%s", dbPort))
	err := ct.dbProxy.Start()
	Expect(err).To(BeNil(), "Could not start DB proxy")
	ct.cfg.DBPort = dbProxyPort
	ct.cfg.DBMaxIdleConnections = 0
}

func (ct *cacheTest) teardown() {
	e2e_common.ServerTeardown(ct.server)
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

func (ct *cacheTest) expectNoResource() {
	expectResponseVal(
		ct.lastResource,
		[]string{},
		func(cifv model.ComponentInstanceFilterValue) []*string {
			return cifv.ServiceCcrn.Values
		})
}

func (ct *cacheTest) addDbResource() {
	ct.addedSeedCollection = ct.seeder.SeedDbWithNFakeData(addResourceCount)
}

func (ct *cacheTest) removeAllResources() {
	Expect(ct.seeder.Clear()).To(BeNil(), "Database Seeder could not clear DB")
}

func (ct *cacheTest) queryResource() {
	ct.lastResource = queryComponentInstanceFilter(
		ct.cfg.Port,
		"../api/graphql/graph/queryCollection/componentInstanceFilter/serviceCcrn.graphqls",
	)
}

func (ct *cacheTest) expectMissHitCounter(expectedMiss, expectedHit int64) {
	stat := ct.server.GetApp().GetCache().GetStat()
	Expect(stat.Miss).To(Equal(expectedMiss))
	Expect(stat.Hit).To(Equal(expectedHit))
}

func missTest(config cacheConfig) {
	// GIVEN Cache is configured
	ct := newCacheTest(config.valkeyUrl, defaultTtl, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval, noProxy)
	defer ct.teardown()

	// WHEN Test resource is queried for the first time
	ct.testResourceIsQueried()

	// THEN Miss = 1 and Hit = 0
	ct.expectMissHitCounter(1, 0)
}

func hitTest(config cacheConfig) {
	// GIVEN Cache is configured
	ct := newCacheTest(config.valkeyUrl, defaultTtl, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval, noProxy)
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
	ct := newCacheTest(config.valkeyUrl, ttl1Millisecond, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval, noProxy)
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
	ct := newCacheTest(config.valkeyUrl, defaultTtl, noConcurrentLimit, noThrottleIntervalMSec, noThrottlePerInterval, noProxy)
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
	ct := newCacheTest(config.valkeyUrl, defaultTtl, noConcurrentLimit, 1000, 2, noProxy)
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
	ct := newCacheTest(config.valkeyUrl, defaultTtl, 2, noThrottleIntervalMSec, noThrottlePerInterval, withProxy)
	defer ct.teardown()

	// AND Resource is queried to update cache with test resource on miss
	ct.queryResource()
	ct.expectMissHitCounter(1, 0)

	// AND DB access is paused for new connections
	ct.dbProxy.HoldNewIncomingConnections()

	// AND Two times resource is queried to exhaust the limit of concurrent background updates
	ct.queryResource()
	ct.queryResource()

	// AND Wait for a moment to start background updates
	time.Sleep(backgroundUpdateStartTimeToWait)

	// AND DB access is allowed for new connections
	ct.dbProxy.DoNotHoldNewIncomingConnections()

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

	// GIVEN resume all hanged DB connections
	ct.dbProxy.ResumeHeldConnections()

	// AND Wait for a moment for resumed background cache update
	time.Sleep(backgroundUpdateTimeToWait)

	// AND Resource is queried
	ct.queryResource()

	// AND Resource contains test resource and added resource due to finished resumed connections
	ct.expectTestResourceAndAddedResource()

	// AND Miss = 1 and Hit = 5
	ct.expectMissHitCounter(1, 5)

	// WHEN All DB resources are removed
	ct.removeAllResources()

	// AND Resource is queried
	ct.queryResource()

	// AND Miss = 1 and Hit = 6
	ct.expectMissHitCounter(1, 6)

	// AND Wait for a moment for background cache update
	time.Sleep(backgroundUpdateTimeToWait)

	// AND Resource is queried
	ct.queryResource()

	// THEN Resource contains no resource
	ct.expectNoResource()

	// AND Miss = 1 and Hit = 7
	ct.expectMissHitCounter(1, 7)
}

var _ = Describe("Using Valkey cache", Label("e2e", "ValkeyCache"), Label("e2e", "Cache"), func() {
	Describe("Check miss", func() {
		It("Should increase miss counter when resource is queried for the first time", func() {
			missTest(valkeyCacheTestConfig)
		})
	})

	Describe("Check hit", func() {
		It("Should increase hit counter when the same resource is queried for the second time", func() {
			hitTest(valkeyCacheTestConfig)
		})
	})

	Describe("Check expired", func() {
		It("Should increase miss counter when the same resource is queried after TTL duration", func() {
			expiredTest(valkeyCacheTestConfig)
		})
	})

	Describe("Check background update", func() {
		It("Should update resource in background on hit so next hit will show new data", func() {
			backgroundUpdateTest(valkeyCacheTestConfig)
		})
	})

	Describe("Check background update rate-based limit", func() {
		It("Should skip background updates above the rate limit AND execute new background updates after interval", func() {
			backgroundUpdateRateBasedLimitTest(valkeyCacheTestConfig)
		})
	})

	Describe("Check background update concurrent-based limit", func() {
		It("Should skip background updates above the concurrent limit AND execute new background updates when first background updates are over", func() {
			backgroundUpdateConcurrentLimitTest(valkeyCacheTestConfig)
		})
	})
})

var _ = Describe("Using In memory cache", Label("e2e", "InMemoryCache"), Label("e2e", "Cache"), func() {
	Describe("Check miss", func() {
		It("Should increase miss counter when resource is queried for the first time", func() {
			missTest(inMemoryCacheTestConfig)
		})
	})

	Describe("Check hit", func() {
		It("Should increase hit counter when the same resource is queried for the second time", func() {
			hitTest(inMemoryCacheTestConfig)
		})
	})

	Describe("Check expired", func() {
		It("Should increase miss counter when the same resource is queried after TTL duration", func() {
			expiredTest(inMemoryCacheTestConfig)
		})
	})

	Describe("Check background update", func() {
		It("Should update resource in background on hit so next hit will show new data", func() {
			backgroundUpdateTest(inMemoryCacheTestConfig)
		})
	})

	Describe("Check background update rate-based limit", func() {
		It("Should skip background updates above the rate limit AND execute new background updates after interval", func() {
			backgroundUpdateRateBasedLimitTest(inMemoryCacheTestConfig)
		})
	})

	Describe("Check background update concurrent-based limit", func() {
		It("Should skip background updates above the concurrent limit AND execute new background updates when first background updates are over", func() {
			backgroundUpdateConcurrentLimitTest(inMemoryCacheTestConfig)
		})
	})
})
