// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/cloudoperators/heureka/scanner/nvd/models"
	p "github.com/cloudoperators/heureka/scanner/nvd/processor"
	s "github.com/cloudoperators/heureka/scanner/nvd/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Also report methods
	log.SetReportCaller(true)
}

func startTimeWindow(scanner *s.Scanner, processor *p.Processor, config s.Config) error {

	startTime, err := time.Parse("2006-01-02", config.StartDate)

	absoluteEnd := time.Now()
	if config.EndDate != "" {
		absoluteEnd, err = time.Parse("2006-01-02", config.EndDate)
	}

	if err != nil {
		return err
	}

	endTime := startTime.AddDate(0, 2, 0)

	for endTime.Before(absoluteEnd) {
		startYear, startMonth, startDay := startTime.Date()
		endYear, endMonth, endDay := endTime.Date()
		start := fmt.Sprintf("%d-%02d-%02dT23:59:59.000", startYear, startMonth, startDay)
		end := fmt.Sprintf("%d-%02d-%02dT23:59:59.000", endYear, endMonth, endDay)

		scanAndProcess(scanner, processor, start, end)

		startTime = startTime.AddDate(0, 2, 0)
		endTime = endTime.AddDate(0, 2, 0)
	}
	return nil
}

func scanAndProcess(scanner *s.Scanner, processor *p.Processor, yesterday string, today string) {
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

func main() {
	var err error
	var scannerCfg s.Config
	err = envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Couldn't initialize scanner config")
	}
	scanner := s.NewScanner(scannerCfg)

	var processorCfg p.Config
	err = envconfig.Process("heureka", &processorCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't configure new processor")
	}

	processor := p.NewProcessor(processorCfg)
	err = processor.Setup()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't setup new processor")
	}

	if scannerCfg.StartDate != "" {
		err = startTimeWindow(scanner, processor, scannerCfg)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't fetch CVEs for time window")
		}
	} else {
		t := time.Now()
		yearToday, monthToday, dayToday := time.Now().Date()
		today := fmt.Sprintf("%d-%02d-%02dT23:59:59.000", yearToday, monthToday, dayToday)

		yearYesterday, monthYesterday, dayYesterday := t.AddDate(0, 0, -2).Date()
		yesterday := fmt.Sprintf("%d-%02d-%02dT00:00:00.000", yearYesterday, monthYesterday, dayYesterday)

		scanAndProcess(scanner, processor, yesterday, today)
	}

}
