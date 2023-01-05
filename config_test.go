// Copyright 2023 Infrable. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func InitTestConfigEnv() {
	os.Setenv("KUBECOST_EXPORTER_CONFIG_NAME", "config")
	os.Setenv("KUBECOST_EXPORTER_CONFIG_TYPE", "yaml")
	os.Setenv("KUBECOST_EXPORTER_CONFIG_PATH", "test")
}

func TestGetAdditionalConfigFromEnv(t *testing.T) {
	t.Run("get additional config from environment", func(t *testing.T) {
		InitTestConfigEnv()
		cn, ct, cp := GetAdditionalConfigFromEnv()
		assert.NotEmpty(t, cn)
		assert.NotEmpty(t, ct)
		assert.NotEmpty(t, cp)
	})
}

func TestNewConfig(t *testing.T) {
	t.Run("default config file exists", func(t *testing.T) {
		assert.NotPanics(t, func() { NewConfig() },
			"The default config file (default.yaml) should be located in the configs/ directory.")
	})
	t.Run("additional config file specified, but not found", func(t *testing.T) {
		InitTestConfigEnv()
		// Config name set, but file does not exist.
		os.Setenv("KUBECOST_EXPORTER_CONFIG_NAME", "not-found")
		c, err := NewConfig()
		assert.NotNil(t, c)
		assert.ErrorContains(t, err, "Additional config not found in 'test/not-found.yaml'")
	})
	t.Run("additional config file specified and exists", func(t *testing.T) {
		InitTestConfigEnv()
		c, err := NewConfig()
		assert.NotNil(t, c)
		assert.NoError(t, err)
	})
}
