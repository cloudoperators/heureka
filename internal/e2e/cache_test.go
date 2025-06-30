// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	"time"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

const (
	ttl24HoursInMSec  = 24 * 60 * 60 * 1000
	testResourceCount = 1
	addResourceCount  = 1
)

var shortTtlTimeToWait = 10 * time.Millisecond
var backgroundUpdateTimeToWait = 100 * time.Millisecond

type cacheTest struct {
	seeder              *test.DatabaseSeeder
	server              *server.Server
	cfg                 util.Config
	db                  *mariadb.SqlDatabase
	seedCollection      *test.SeedCollection
	lastResource        model.ComponentInstanceFilterValue
	addedSeedCollection *test.SeedCollection
}

func newCacheTest(redisUrl string, ttlMSec int64) *cacheTest {
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

	ct.seedCollection = ct.seeder.SeedDbWithNFakeData(testResourceCount)
	return &ct
}

func newRedisCacheTest(ttlMSec int64) *cacheTest {
	return newCacheTest("localhost:6379", ttlMSec)
}

func newInMemoryCacheTest(ttlMSec int64) *cacheTest {
	return newCacheTest("", ttlMSec)
}

func newNoCacheTest() *cacheTest {
	return newCacheTest("", 0)
}

func (ct *cacheTest) teardown() {
	ct.server.BlockingStop()
	dbm.TestTearDown(ct.db)
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

var _ = Describe("Using Redis cache", Label("e2e", "RedisCache"), Label("e2e", "Cache"), func() {
	Describe("Check miss", func() {
		Context("Redis cache is configured", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newRedisCacheTest(ttl24HoursInMSec)
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Test resource is queried for the first time", func() {
				BeforeEach(func() {
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Miss counter should be equal 1 and Hit counter should be equal to 0", func() {
					ct.expectMissHitCounter(1, 0)
				})
			})
		})
	})
	Describe("Check hit", func() {
		Context("Redis cache is configured and cache contains test resource", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newRedisCacheTest(ttl24HoursInMSec)
				ct.queryResource()
				ct.expectTestResource()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Cached test resource is queried", func() {
				BeforeEach(func() {
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Hit counter should be equal 1 and Miss counter should be equal to 1", func() {
					ct.expectMissHitCounter(1, 1)
				})
			})
		})
	})
	Describe("Check expired", func() {
		Context("Redis cache is configured with short TTL duration and test resource is queried for the first time", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newRedisCacheTest(1)
				ct.queryResource()
				ct.expectTestResource()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("TTL duration elapsed and test resource is queried", func() {
				BeforeEach(func() {
					time.Sleep(shortTtlTimeToWait)
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Miss counter should be equal to 2 and Hit counter should be equal 0", func() {
					ct.expectMissHitCounter(2, 0)
				})
			})
		})
	})
	Describe("Check background update", func() {
		Context("Redis cache is configured and test resource is queried for the first time and added extra DB resource", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newRedisCacheTest(ttl24HoursInMSec)
				ct.queryResource()
				ct.expectTestResource()
				ct.addDbResource()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Resource is queried", func() {
				BeforeEach(func() {
					ct.queryResource()
				})
				It("Resource contains test resource and Miss counter should be equal to 1 and Hit counter should be equal 1", func() {
					ct.expectTestResource()
					ct.expectMissHitCounter(1, 1)
				})
				Context("Wait for a moment for background cache update", func() {
					BeforeEach(func() {
						time.Sleep(backgroundUpdateTimeToWait)
					})
					When("Resource is queried", func() {
						BeforeEach(func() {
							ct.queryResource()
						})
						It("Resource contains test resource and added resource and Miss counter should be equal to 1 and Hit counter should be equal 2", func() {
							ct.expectTestResourceAndAddedResource()
							ct.expectMissHitCounter(1, 2)
						})
					})
				})
			})
		})
	})
})

var _ = Describe("Using In memory cache", Label("e2e", "InMemoryCache"), Label("e2e", "Cache"), func() {
	Describe("Check miss", func() {
		Context("In memory cache is configured", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newInMemoryCacheTest(ttl24HoursInMSec)
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Test resource is queried for the first time", func() {
				BeforeEach(func() {
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Miss counter should be equal 1 and Hit counter should be equal to 0", func() {
					ct.expectMissHitCounter(1, 0)
				})
			})
		})
	})
	Describe("Check hit", func() {
		Context("In memory cache is configured and cache contains test resource", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newInMemoryCacheTest(ttl24HoursInMSec)
				ct.queryResource()
				ct.expectTestResource()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Cached test resource is queried", func() {
				BeforeEach(func() {
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Hit counter should be equal 1 and Miss counter should be equal to 1", func() {
					ct.expectMissHitCounter(1, 1)
				})
			})
		})
	})
	Describe("Check expired", func() {
		Context("In memory cache is configured with short TTL duration and test resource is queried for the first time", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newInMemoryCacheTest(1)
				ct.queryResource()
				ct.expectTestResource()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("TTL duration elapsed and test resource is queried", func() {
				BeforeEach(func() {
					time.Sleep(shortTtlTimeToWait)
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Miss counter should be equal to 2 and Hit counter should be equal 0", func() {
					ct.expectMissHitCounter(2, 0)
				})
			})
		})
	})
	Describe("Check background update", func() {
		Context("In memory cache is configured and test resource is queried for the first time and added extra DB resource", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newInMemoryCacheTest(ttl24HoursInMSec)
				ct.queryResource()
				ct.expectTestResource()
				ct.addDbResource()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Resource is queried", func() {
				BeforeEach(func() {
					ct.queryResource()
				})
				It("Resource contains test resource and Miss counter should be equal to 1 and Hit counter should be equal 1", func() {
					ct.expectTestResource()
					ct.expectMissHitCounter(1, 1)
				})
				Context("Wait for a moment for background cache update", func() {
					BeforeEach(func() {
						time.Sleep(backgroundUpdateTimeToWait)
					})
					When("Resource is queried", func() {
						BeforeEach(func() {
							ct.queryResource()
						})
						It("Resource contains test resource and added resource and Miss counter should be equal to 1 and Hit counter should be equal 2", func() {
							ct.expectTestResourceAndAddedResource()
							ct.expectMissHitCounter(1, 2)
						})
					})
				})
			})
		})
	})
})

var _ = Describe("Using No cache", Label("e2e", "NoCache"), Label("e2e", "Cache"), func() {
	Describe("Check miss", func() {
		Context("No cache is configured", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newNoCacheTest()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Test resource is queried for the first time", func() {
				BeforeEach(func() {
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Miss counter should be equal 0 and Hit counter should be equal to 0", func() {
					ct.expectMissHitCounter(0, 0)
				})
			})
		})
	})
	Describe("Check hit", func() {
		Context("No cache is configured and cache contains test resource", func() {
			var ct *cacheTest
			BeforeEach(func() {
				ct = newNoCacheTest()
				ct.queryResource()
				ct.expectTestResource()
			})
			AfterEach(func() {
				ct.teardown()
			})
			When("Cached test resource is queried", func() {
				BeforeEach(func() {
					ct.queryResource()
					ct.expectTestResource()
				})
				It("Hit counter should be equal 0 and Miss counter should be equal to 0", func() {
					ct.expectMissHitCounter(0, 0)
				})
			})
		})
	})
})
