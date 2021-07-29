/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package logging

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"iam/pkg/config"
)

var loggerInitOnce sync.Once

// use zap for better performance
var apiLogger *zap.Logger
var webLogger *zap.Logger

// use logrus for better usage
var sqlLogger *logrus.Logger
var auditLogger *logrus.Logger
var componentLogger *logrus.Logger

// InitLogger ...
func InitLogger(logger *config.Logger) {
	initSystemLogger(&logger.System)

	loggerInitOnce.Do(func() {
		// json logger
		apiLogger = newZapJSONLogger(&logger.API)
		webLogger = newZapJSONLogger(&logger.Web)

		sqlLogger = newJSONLogger(&logger.SQL)
		auditLogger = newJSONLogger(&logger.Audit)
		componentLogger = newJSONLogger(&logger.Component)
	})
}

func initSystemLogger(cfg *config.LogConfig) {
	writer, err := getWriter(cfg.Writer, cfg.Settings)
	if err != nil {
		panic(err)
	}
	// 日志输出到stdout
	logrus.SetOutput(writer)
	// 设置日志格式, 不需要颜色
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	DisableColors:   true,
	// 	FullTimestamp:   true,
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// })
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})

	// 设置日志级别
	l, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		fmt.Println("system logger settings level invalid, will use level: info")
		l = logrus.InfoLevel
		// panic(err)
	}
	logrus.SetLevel(l)
	// https://github.com/sirupsen/logrus#logging-method-name
	// DONT OPEN IT
	// 显示代码行数
	// logrus.SetReportCaller(true)
}

func newJSONLogger(cfg *config.LogConfig) *logrus.Logger {
	jsonLogger := logrus.New()

	writer, err := getWriter(cfg.Writer, cfg.Settings)
	if err != nil {
		panic(err)
	}
	jsonLogger.SetOutput(writer)

	// apiLogger.SetFormatter(&logrus.JSONFormatter{
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// })
	jsonLogger.SetFormatter(&JSONFormatter{})
	// 设置日志级别
	l, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		fmt.Println("api logger settings level invalid, will use level: info")
		l = logrus.InfoLevel
		// panic(err)
	}
	jsonLogger.SetLevel(l)

	return jsonLogger
}

// parseZapLogLevel takes a string level and returns the zap log level constant.
func parseZapLogLevel(lvl string) (zapcore.Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return zap.PanicLevel, nil
	case "fatal":
		return zap.FatalLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	case "warn", "warning":
		return zap.WarnLevel, nil
	case "info":
		return zap.InfoLevel, nil
	case "debug":
		return zap.DebugLevel, nil
	}

	var l zapcore.Level
	return l, fmt.Errorf("not a valid logrus Level: %q", lvl)
}

func newZapJSONLogger(cfg *config.LogConfig) *zap.Logger {
	writer, err := getWriter(cfg.Writer, cfg.Settings)
	if err != nil {
		panic(err)
	}
	w := zapcore.AddSync(writer)

	// 设置日志级别
	l, err := parseZapLogLevel(cfg.Level)
	if err != nil {
		fmt.Println("api logger settings level invalid, will use level: info")
		l = zap.InfoLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		l,
	)
	return zap.New(core)
}

// GetAPILogger api log
func GetAPILogger() *zap.Logger {
	// if not init yet, use system logger
	if apiLogger == nil {
		apiLogger, _ = zap.NewProduction()
		defer apiLogger.Sync()
	}

	return apiLogger
}

// GetWebLogger web log
func GetWebLogger() *zap.Logger {
	// if not init yet, use system logger
	if webLogger == nil {
		webLogger, _ = zap.NewProduction()
		defer webLogger.Sync()
	}
	return webLogger
}

// GetSQLLogger sql log
func GetSQLLogger() *logrus.Logger {
	// if not init yet, use system logger
	if sqlLogger == nil {
		return logrus.StandardLogger()
	}
	return sqlLogger
}

// GetAuditLogger audit log
func GetAuditLogger() *logrus.Logger {
	// if not init yet, use system logger
	if auditLogger == nil {
		return logrus.StandardLogger()
	}
	return auditLogger
}

// GetSystemLogger ...
func GetSystemLogger() *logrus.Logger {
	return logrus.StandardLogger()
}

// GetComponentLogger ...
func GetComponentLogger() *logrus.Logger {
	// if not init yet, use system logger
	if componentLogger == nil {
		return logrus.StandardLogger()
	}
	return componentLogger
}
