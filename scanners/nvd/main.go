// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"
	"os"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/scanners/nvd/models"
	"github.wdf.sap.corp/cc/heureka/scanners/nvd/processor"
	"github.wdf.sap.corp/cc/heureka/scanners/nvd/scanner"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

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
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't get CVEs")
	}

	var processorCfg processor.Config
	err = envconfig.Process("heureka", &processorCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't configure new processor")
	}

	processor := processor.NewProcessor(processorCfg)
	err = processor.Setup()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't setup new processor")
	}

	for _, cve := range cves {
		err = processor.Process(&cve.Cve)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"CVEID": &cve.Cve.Id,
			}).Warn("Couldn't process CVE")
		}
	}
}
