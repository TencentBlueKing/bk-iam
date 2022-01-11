/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Register mysql resource
	_ "github.com/go-sql-driver/mysql"

	// init debug entry pool
	_ "iam/pkg/logging/debug"

	"iam/pkg/server"
)

// cmd for iam
var cfgFile string

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yml;required)")
	rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")

	rootCmd.MarkFlagRequired("config")
	viper.SetDefault("author", "blueking-paas")
}

var rootCmd = &cobra.Command{
	Use:   "bk-iam",
	Short: "bi-iam is Identity and Access Management System",
	Long: `BlueKing Identity and Access Management (BK-IAM)
           is a service that helps you securely control access to system resources`,

	Run: func(cmd *cobra.Command, args []string) {
		Start()
	},
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Start ...
func Start() {
	fmt.Println("It's IAM")

	// init rand
	rand.Seed(time.Now().UnixNano())

	// 0. init config
	if cfgFile != "" {
		// Use config file from the flag.
		log.Infof("Load config file: %s", cfgFile)
		viper.SetConfigFile(cfgFile)
	}
	initConfig()

	if globalConfig.Debug {
		fmt.Println(globalConfig)
	}

	// 1. init
	initLogger()
	initSentry()
	initMetrics()
	initDatabase()
	initRedis()
	// NOTE: should be after initRedis
	initCaches()
	initPolicyCacheSettings()
	initSuperAppCode()
	initSuperUser()
	initSupportShieldFeatures()
	initComponents()
	initQuota()
	initSwitch()

	// 2. watch the signal
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		interrupt(cancelFunc)
	}()

	// 2. start expression clean task
	initExpressionCleanTask(ctx)

	// 3. start the server
	httpServer := server.NewServer(globalConfig)
	httpServer.Run(ctx)
}

// a context canceled when SIGINT or SIGTERM are notified
func interrupt(onSignal func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	for s := range c {
		log.Infof("Caught signal %s. Exiting.", s)
		onSignal()
		close(c)
	}
}
