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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"iam/pkg/task"
)

func init() {
	transferCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yml;required)")
	transferCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")

	transferCmd.MarkFlagRequired("config")
}

// transferCmd represents asynchronous task transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "bk-iam transfer is asynchronous task transfer",
	Long: `BlueKing Identity and Access Management (BK-IAM)
		transfer is used to transform subject action alter event`,
	Run: func(cmd *cobra.Command, args []string) {
		StartTransfer()
	},
}

// StartTransfer ...
func StartTransfer() {
	fmt.Println("It's IAM transfer")

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
	initWorker()

	// 2. watch the signal
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		interrupt(cancelFunc)
	}()

	// 3. start sync transfer
	transfer := task.NewTransfer()
	transfer.Run(ctx)
}

func init() {
	rootCmd.AddCommand(transferCmd)
}
