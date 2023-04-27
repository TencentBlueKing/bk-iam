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
	"fmt"
	"iam/pkg/database"
	"iam/pkg/service"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	clearCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yml;required)")
	clearCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")

	clearCmd.MarkFlagRequired("config")
}

// checkerCmd represents asynchronous task check command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "bk-iam checker is asynchronous task checker",
	Long: `BlueKing Identity and Access Management (BK-IAM)
			checker is used to check asynchronous task`,
	Run: func(cmd *cobra.Command, args []string) {
		StartClear()
	},
}

// StartChecker ...
func StartClear() {
	fmt.Println("It's IAM clear")

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

	svc := service.NewModelChangeService()
	policyService := service.NewPolicyService()
	actionService := service.NewActionService()

	events, err := svc.ListByStatus(service.ModelChangeEventStatusFinished, 50)
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		if event.Type != service.ModelChangeEventTypeActionPolicyDeleted || event.ModelType != service.ModelChangeEventModelTypeAction {
			continue
		}

		// check action pk not exists
		thinActions, err := actionService.ListThinActionByPKs([]int64{event.ModelPK})
		if err != nil {
			panic(err)
		}

		if len(thinActions) > 0 {
			continue
		}

		// 2. delete event task
		tx, err := database.GenerateDefaultDBTx()
		if err != nil {
			panic(err)
		}
		defer database.RollBackWithLog(tx)

		// 2. 删除abac policy
		err = policyService.DeleteByActionPKWithTx(tx, event.ModelPK)
		if err != nil {
			panic(err)
		}

		err = tx.Commit()
		if err != nil {
			panic(err)
		}

	}

}

func init() {
	rootCmd.AddCommand(clearCmd)
}
