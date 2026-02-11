// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type envLogConfig struct {
	PrettyPrint string `envconfig:"LOG_PRETTY_PRINT" default:"false" json:"log_pretty_print"`
	Format      string `envconfig:"LOG_FORMAT" default:"json" json:"log_format"`
	Location    string `envconfig:"LOG_LOCATION" default:"stdout" json:"log_location"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"debug" json:"log_level"`
}

type LogConfig struct {
	Formatter   logrus.Formatter
	Writer      io.Writer
	PrettyPrint bool
	LogLevel    logrus.Level
}

func (l *LogConfig) SetFormatter(v string) {
	switch v {
	case "json":
		l.Formatter = &logrus.JSONFormatter{
			PrettyPrint: l.PrettyPrint,
			// CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			//	_, filename := path.Split(f.File)
			//	return f.Function, fmt.Sprintf("%s: %d", filename, f.Line)
			// },
		}

	case "text":
		l.Formatter = &logrus.TextFormatter{}
	default:
		logrus.Warn(fmt.Sprintf("Unkown LOG_FORMAT provided: %s, Using default: %s", v, "json"))
	}
}

func (l *LogConfig) SetPrettyPrint(v string) {
	switch v {
	case "true":
		l.PrettyPrint = true
	case "false":
		l.PrettyPrint = false
	default:
		logrus.Warn(fmt.Sprintf("Unkown LOG_FORMAT provided: %s, Using default: %s", v, "json"))
	}
}

func (l *LogConfig) SetWriter(v string) {
	switch v {
	case "stdout":
		l.Writer = os.Stdout
	default:
		f, err := os.OpenFile("/var/log/heureka.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			logrus.Warn(fmt.Sprintf("Error while creating log io.Writer for file: %s, Using default: %s", v, "stdout"))
			l.Writer = os.Stdout
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("error during file closing: %s", err)
			}
		}()

		l.Writer = f
	}
}

func (l *LogConfig) SetLogLevel(v string) {
	switch v {
	case "panic":
		l.LogLevel = logrus.PanicLevel
	case "fatal":
		l.LogLevel = logrus.FatalLevel
	case "error":
		l.LogLevel = logrus.ErrorLevel
	case "warn":
		l.LogLevel = logrus.WarnLevel
	case "info":
		l.LogLevel = logrus.InfoLevel
	case "debug":
		l.LogLevel = logrus.DebugLevel
	case "trace":
		l.LogLevel = logrus.TraceLevel
	default:
		logrus.Warn(fmt.Sprintf("Unkown LOG_LEVEL defined: %s, Using default: %s", v, "debug"))
		l.LogLevel = logrus.DebugLevel
	}
}

func GetLogConfig() *LogConfig {
	var envCfg envLogConfig
	err := envconfig.Process("logConfig", &envCfg)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("Failure while reading env config for log configuration: %s", err))
	}

	logConfig := &LogConfig{}
	logConfig.SetPrettyPrint(envCfg.PrettyPrint)
	logConfig.SetFormatter(envCfg.Format)
	logConfig.SetWriter(envCfg.Location)
	logConfig.SetLogLevel(envCfg.LogLevel)

	return logConfig
}
