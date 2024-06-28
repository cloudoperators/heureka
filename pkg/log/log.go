// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/sirupsen/logrus"
)

// Loads Configuration from Environment and configures the default logger
// Relevant LogLevel
func InitLog() {
	cfg := GetLogConfig()
	logrus.SetLevel(cfg.LogLevel)
	logrus.SetFormatter(cfg.Formatter)
	logrus.SetOutput(cfg.Writer)
	logrus.SetReportCaller(true)
}
