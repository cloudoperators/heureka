// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/models"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/processor"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/scanner"
)

func main() {
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		fmt.Println(err)
	}
	scanner := scanner.NewScanner(scannerCfg)

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

	var processorCfg processor.Config
	err = envconfig.Process("heureka", &processorCfg)
	if err != nil {
		fmt.Println(err)
	}

	processor := processor.NewProcessor(processorCfg)

	for _, cve := range cves {
		err = processor.Process(&cve.Cve)
		if err != nil {
			fmt.Println(err)
		}
	}

}
