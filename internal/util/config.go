// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Config struct {
	Port string `envconfig:"SERVER_PORT" default:":80" json:"-"`
	//Regions            []string `envconfig:"REGIONS" required:"true" json:"regions"`
	//CloudAdminUsername string   `envconfig:"OS_USERNAME" required:"true" json:"cloudAdminUser"`
	//CloudAdminPassword string   `envconfig:"OS_PASSWORD" required:"true" json:"-"`
	DBAddress            string `envconfig:"DB_ADDRESS" required:"true" json:"dbAddress"`
	DBUser               string `envconfig:"DB_USER" required:"true" json:"dbUser"`
	DBPassword           string `envconfig:"DB_PASSWORD" required:"true" json:"-"`
	DBRootPassword       string `envconfig:"DB_ROOT_PASSWORD" required:"true" json:"-"`
	DBPort               string `envconfig:"DB_PORT" required:"true" json:"dbPort"`
	DBName               string `envconfig:"DB_NAME" required:"true" json:"dbDbName"`
	DBSchema             string `envconfig:"DB_SCHEMA" required:"true" json:"dbSchema"`
	DBMaxIdleConnections int    `envconfig:"DB_MAX_IDLE_CONNECTIONS" default:"10" json:"dBMaxIdleConnections"`
	DBMaxOpenConnections int    `envconfig:"DB_MAX_OPEN_CONNECTIONS" default:"100" json:"dbMaxOpenConnections"`
	//VasApiAddress      string   `envconfig:"VAS_API_ADDRESS" required:"true" json:"vasApiAddress"`
	//VasApiToken        string   `envconfig:"VAS_API_TOKEN" required:"true" json:"-"`
	//NvdApiToken        string   `envconfig:"NVD_API_TOKEN" required:"true" json:"-"`
	//OidcClientId       string   `envconfig:"OIDC_CLIENT_ID" required:"true" json:"-"`
	//OidcUrl            string   `envconfig:"OIDC_URL" required:"true" json:"-"`
	//Environment        string   `envconfig:"ENVIRONMENT" required:"true" json:"environment"`
	//// https://pkg.go.dev/github.com/robfig/cron#hdr-Predefined_schedules
	//DiscoverySchedule string `envconfig:"DISOVERY_SCHEDULE" default:"0 0 0 * * *" json:"discoverySchedule"`
	SeedMode bool `envconfig:"SEED_MODE" required:"false" default:"false" json:"seedMode"`
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
		{"Database Name", c.DBSchema},
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
	table.SetHeader([]string{"Variable", "Value"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(data)
	table.Render()
}

const HeurekaFiglet = `
 _   _                     _         
| | | | ___ _   _ _ __ ___| | ____ _ 
| |_| |/ _ \ | | | '__/ _ \ |/ / _' |
|  _  |  __/ |_| | | |  __/   < (_| |
|_| |_|\___|\__,_|_|  \___|_|\_\__,_|
`
