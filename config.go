// Copyright 2022 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Application configuration.
//
// For documentation on Viper, see the following:
//   - https://github.com/spf13/viper
package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	DefaultConfigName = "default"
	DefaultConfigType = "yaml"
	DefaultConfigPath = "./configs"
)

const (
	AdditionalConfigNameEnvVar = "KUBECOST_EXPORTER_CONFIG_NAME"
	AdditionalConfigTypeEnvVar = "KUBECOST_EXPORTER_CONFIG_TYPE"
	AdditionalConfigPathEnvVar = "KUBECOST_EXPORTER_CONFIG_PATH"
)

// Get additional configuration values (name, type, path) from environment.
func GetAdditionalConfigFromEnv() (string, string, string) {
	return os.Getenv(AdditionalConfigNameEnvVar),
		os.Getenv(AdditionalConfigTypeEnvVar),
		os.Getenv(AdditionalConfigPathEnvVar)
}

// Initialize new Viper configuration.
//
// See configs/default.yaml for default configuration values.
//
// The default config file should be located in the configs/ directory.
//
// Additional configuration values can be specified by the file formed from the
// following environment variables:
//   - KUBECOST_EXPORTER_CONFIG_NAME
//   - KUBECOST_EXPORTER_CONFIG_TYPE
//   - KUBECOST_EXPORTER_CONFIG_PATH
func NewConfig() (*viper.Viper, error) {
	v0, v1 := viper.New(), viper.New()
	v0.SetConfigName(DefaultConfigName)
	v0.SetConfigType(DefaultConfigType)
	v0.AddConfigPath(DefaultConfigPath)
	err := v0.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}
	if cn, ct, cp := GetAdditionalConfigFromEnv(); cn != "" && ct != "" && cp != "" {
		v1.SetConfigName(cn)
		v1.SetConfigType(ct)
		v1.AddConfigPath(cp)
		err = v1.ReadInConfig()
		if err != nil {
			err = fmt.Errorf("Additional config not found in '%s/%s.%s'", cp, cn, ct)
		}
	}
	// MergeConfigMap merges the configuration from the map given with an
	// existing config.
	v0.MergeConfigMap(v1.AllSettings())
	return v0, err
}
