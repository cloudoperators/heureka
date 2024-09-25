// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

type Config struct {
	StartDate string `envconfig:"NVD_START_DATE" default:"" json:"-"`
	EndDate   string `envconfig:"NVD_END_DATE" default:"" json:"-"`
	NvdApiUrl string `envconfig:"NVD_API_URL" required:"true" json:"-"`
	NvdApiKey string `envconfig:"NVD_API_KEY" required:"true" json:"-"`
	// default value and maximum allowable limit is 2,000
	NvdResultsPerPage string `envconfig:"NVD_RESULTS_PER_PAGE" default:"2000" json:"-"`
}
