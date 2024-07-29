// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/models"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/processor"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/scanner"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/util"
	"golang.org/x/time/rate"
)

func main() {
	var cfg util.Config
	err := envconfig.Process("heureka", &cfg)
	if err != nil {
		fmt.Println(err)
	}
	// The public rate limit (without an API key) is 5 requests in a rolling 30 second window; the rate limit with an API key is 50 requests in a rolling 30 second window
	rl := rate.NewLimiter(rate.Every(30*time.Second/50), 50)
	scanner := scanner.NewScanner(cfg.NvdApiUrl, cfg.NvdApiKey, cfg.NvdResultsPerPage, rl)

	t := time.Now()
	yearToday, monthToday, dayToday := time.Now().Date()
	today := fmt.Sprintf("%d-%02d-%02dT23:59:59.000", yearToday, monthToday, dayToday)
	yearYesterday, monthYesterday, dayYesterday := t.AddDate(0, 0, -1).Date()
	yesterday := fmt.Sprintf("%d-%02d-%02dT00:00:00.000", yearYesterday, monthYesterday, dayYesterday)

	filter := models.CveFilter{
		PubStartDate: yesterday,
		PubEndDate:   today,
	}

	cves, err := scanner.GetCVEs(filter)

	if err != nil {
		fmt.Println(err)
	}

	processor := processor.NewProcessor(cfg.HeurekaUrl)

	for _, cve := range cves {
		err = processor.Process(&cve.Cve)
		if err != nil {
			fmt.Println(err)
		}
	}

}
