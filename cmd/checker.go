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
	"time"

	"iam/pkg/task"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// checkerCmd represents asynchronous task check command
var checkerCmd = &cobra.Command{
	Use:   "checker",
	Short: "bk-iam checker is asynchronous task checker",
	Long: `BlueKing Identity and Access Management (BK-IAM)
		   worker is used to check asynchronous task`,
	Run: func(cmd *cobra.Command, args []string) {
		StartChecker()
	},
}

// StartChecker ...
func StartChecker() {
	fmt.Println("It's IAM worker")

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
	initRmqCleaner()

	// 2. watch the signal
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		interrupt(cancelFunc)
	}()

	// 3. start sync worker
	checker := task.NewChecker()
	checker.Run(ctx)
}

func init() {
	rootCmd.AddCommand(checkerCmd)
}
