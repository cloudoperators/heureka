// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

type Config struct {
	Port string `envconfig:"SERVER_PORT" default:"80" json:"-"`
	//Regions            []string `envconfig:"REGIONS" required:"true" json:"regions"`
	//CloudAdminUsername string   `envconfig:"OS_USERNAME" required:"true" json:"cloudAdminUser"`
	//CloudAdminPassword string   `envconfig:"OS_PASSWORD" required:"true" json:"-"`
	DBAddress            string `envconfig:"DB_ADDRESS" required:"true" json:"dbAddress"`
	DBUser               string `envconfig:"DB_USER" required:"true" json:"dbUser"`
	DBPassword           string `envconfig:"DB_PASSWORD" required:"true" json:"-"`
	DBRootPassword       string `envconfig:"DB_ROOT_PASSWORD" required:"true" json:"-"`
	DBPort               string `envconfig:"DB_PORT" required:"true" json:"dbPort"`
	DBName               string `envconfig:"DB_NAME" required:"true" json:"dbDbName"`
	DBMaxIdleConnections int    `envconfig:"DB_MAX_IDLE_CONNECTIONS" default:"10" json:"dBMaxIdleConnections"`
	DBMaxOpenConnections int    `envconfig:"DB_MAX_OPEN_CONNECTIONS" default:"100" json:"dbMaxOpenConnections"`
	DBTrace              bool   `envconfig:"DB_TRACE" default:"false" json:"-"`
	//VasApiAddress              string   `envconfig:"VAS_API_ADDRESS" required:"true" json:"vasApiAddress"`
	//VasApiToken                string   `envconfig:"VAS_API_TOKEN" required:"true" json:"-"`
	//NvdApiToken                string   `envconfig:"NVD_API_TOKEN" required:"true" json:"-"`
	//OidcClientId               string   `envconfig:"OIDC_CLIENT_ID" required:"true" json:"-"`
	//OidcUrl                    string   `envconfig:"OIDC_URL" required:"true" json:"-"`
	//Environment                string   `envconfig:"ENVIRONMENT" required:"true" json:"environment"`
	//// https://pkg.go.dev/github.com/robfig/cron#hdr-Predefined_schedules
	//DiscoverySchedule string `envconfig:"DISOVERY_SCHEDULE" default:"0 0 0 * * *" json:"discoverySchedule"`
	SeedMode                      bool   `envconfig:"SEED_MODE" required:"false" default:"false" json:"seedMode"`
	AuthTokenSecret               string `envconfig:"AUTH_TOKEN_SECRET" required:"false" json:"-"`
	AuthOidcClientId              string `envconfig:"AUTH_OIDC_CLIENT_ID" required:"false" json:"-"`
	AuthOidcUrl                   string `envconfig:"AUTH_OIDC_URL" required:"false" json:"-"`
	DefaultIssuePriority          int64  `envconfig:"DEFAULT_ISSUE_PRIORITY" default:"100" json:"defaultIssuePriority"`
	DefaultRepositoryName         string `envconfig:"DEFAULT_REPOSITORY_NAME" default:"nvd" json:"defaultRepositoryName"`
	CacheEnable                   bool   `envconfig:"CACHE_ENABLE" default:"false" json:"-"`
	CacheValkeyUrl                string `envconfig:"CACHE_VALKEY_URL" default:"" json:"-"`
	CacheValkeyPassword           string `envconfig:"CACHE_VALKEY_PASSWORD" default:"" json:"-"`
	CacheValkeyUsername           string `envconfig:"CACHE_VALKEY_USERNAME" default:"" json:"-"`
	CacheValkeyClientName         string `envconfig:"CACHE_VALKEY_CLIENT_NAME" default:"" json:"-"`
	CacheMonitorMSec              int64  `envconfig:"CACHE_MONITOR_MSEC" default:"0" json:"-"`
	CacheMaxDbConcurrentRefreshes int    `envconfig:"CACHE_MAX_DB_CONCURRENT_REFRESHES" default:"-1" json:"-"`
	CacheThrottleIntervalMSec     int64  `envconfig:"CACHE_THROTTLE_INTERVAL_MSEC" default:"0" json:"-"`
	CacheThrottlePerInterval      int    `envconfig:"CACHE_THROTTLE_PER_INTERVAL" default:"1" json:"-"`
	CpuProfilerFilePath           string `envconfig:"CPU_PROFILER_FILE_PATH" default:"" json:"-"`
}

func (c *Config) ConfigToConsole() {
	data := [][]string{
		{"Port:", c.Port},
		//{"Regions", fmt.Sprintf("%v", c.Regions)},
		//{"CloudAdmin Username", c.CloudAdminUsername},
		//{"CloudAdmin Password", strings.Repeat("*", 10)},
		{"Database Address", c.DBAddress},
		{"Database Port", c.DBPort},
		{"Database Name", c.DBName},
		{"Database Username", c.DBUser},
		{"Database Password", strings.Repeat("*", 10)},
		//{"VAS API Address", c.VasApiAddress},
		//{"Environment", c.Environment},
		//{"Postgres Password", strings.Repeat("*", 10)},
		//{"VAS API Token", strings.Repeat("*", 10)},
		//{"NVD API Token", strings.Repeat("*", 10)},
		//{"OIDC Client Id", strings.Repeat("*", 10)},
		//{"OIDC URL", c.OidcUrl},
		//{"Discovery Schedule", c.DiscoverySchedule},
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Variable", "Value"})
	table.Configure(func(config *tablewriter.Config) {
		config.Row.Formatting.Alignment = tw.AlignLeft
	})
	table.Bulk(data)
	table.Render()
}

const HeurekaFiglet = `
 _   _                     _         
| | | | ___ _   _ _ __ ___| | ____ _ 
| |_| |/ _ \ | | | '__/ _ \ |/ / _' |
|  _  |  __/ |_| | | |  __/   < (_| |
|_| |_|\___|\__,_|_|  \___|_|\_\__,_|
`
