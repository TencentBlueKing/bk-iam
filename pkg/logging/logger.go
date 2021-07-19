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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func getWriter(writerType string, settings map[string]string) (io.Writer, error) {
	switch writerType {
	case "os":
		return getOSWriter(settings)
	case "file":
		return getFileWriter(settings)
	default:
		// return nil, fmt.Errorf("logging writer type not support %s", writerType)
		return getOSWriter(map[string]string{"name": "stdout"})
	}
}

func getOSWriter(settings map[string]string) (io.Writer, error) {
	switch settings["name"] {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return os.Stdout, nil
		// return nil, fmt.Errorf("os writer not support %s", settings["name"])
	}
}

func getFileWriter(settings map[string]string) (io.Writer, error) {
	// path, name, backups, size, age
	path, ok := settings["path"]
	if ok {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("file path %s not exists", path)
		}
	} else {
		return nil, errors.New("log file path should not be empty")
	}

	filename := settings["name"]

	backups := 10
	backupsStr, ok := settings["backups"]
	if ok {
		backupsInt, err := strconv.Atoi(backupsStr)
		if err != nil {
			return nil, errors.New("backups should be integer")
		}
		backups = backupsInt
	}

	size := 50
	sizeStr, ok := settings["size"]
	if ok {
		sizeInt, err := strconv.Atoi(sizeStr)
		if err != nil {
			return nil, errors.New("size should be integer")
		}
		size = sizeInt
	}

	age := 7
	ageStr, ok := settings["age"]
	if ok {
		ageInt, err := strconv.Atoi(ageStr)
		if err != nil {
			return nil, errors.New("age should be integer")
		}
		age = ageInt
	}

	logPath := filename
	if path != "" {
		rawPath := strings.TrimSuffix(path, "/")
		logPath = filepath.Join(rawPath, filename)
	}

	writer := &lumberjack.Logger{
		Filename: logPath,
		// megabytes
		MaxSize:    size,
		MaxBackups: backups,
		// days
		MaxAge:    age,
		LocalTime: true,
	}

	return writer, nil
}
