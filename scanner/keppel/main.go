// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/cloudoperators/heureka/scanners/keppel/processor"
	"github.com/cloudoperators/heureka/scanners/keppel/scanner"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	var processorCfg processor.Config
	err = envconfig.Process("heureka", &processorCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	keppelScanner := scanner.NewScanner(scannerCfg)
	keppelProcessor := processor.NewProcessor(processorCfg)

	err = keppelScanner.Setup()
	if err != nil {
		fmt.Println(err)
		return
	}
	accounts, err := keppelScanner.ListAccounts()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, account := range accounts {
		repositories, err := keppelScanner.ListRepositories(account.Name)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, repository := range repositories {
			component, err := keppelProcessor.ProcessRepository(repository)
			if err != nil {
				fmt.Println(err)
				componentPtr, err := keppelProcessor.GetComponent(repository.Name)
				if err != nil {
					fmt.Println(err)
				}
				component = *componentPtr
			}

			manifests, err := keppelScanner.ListManifests(account.Name, repository.Name)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, manifest := range manifests {
				componentVersion, err := keppelProcessor.ProcessManifest(manifest, component.ID)
				if err != nil {
					fmt.Println(err)
					componentVersionPtr, err := keppelProcessor.GetComponentVersion(manifest.Digest)
					if err != nil {
						fmt.Println(err)
					}
					componentVersion = *componentVersionPtr
				}
				trivyReport, err := keppelScanner.GetTrivyReport(account.Name, repository.Name, manifest.Digest)
				if err != nil {
					fmt.Println(err)
				}
				keppelProcessor.ProcessReport(*trivyReport, componentVersion.ID)
			}
		}
	}
}
