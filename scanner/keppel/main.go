// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

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

	keppelScanner := scanner.NewScanner(scannerCfg)

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
			manifests, err := keppelScanner.ListManifests(account.Name, repository.Name)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, manifest := range manifests {
				keppelScanner.GetTrivyReport(account.Name, repository.Name, manifest.Digest)
			}
		}
	}
}
