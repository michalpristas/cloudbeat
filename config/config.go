// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/config"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
)

const DefaultNamespace = "default"

const ResultsDatastreamIndexPrefix = "logs-cloud_security_posture.findings"

const (
	InputTypeVanillaK8s = "cloudbeat/cis_k8s"
	InputTypeEks        = "cloudbeat/cis_eks"
)

type Fetcher struct {
	Name string `config:"name"` // Name of the fetcher
}

type Config struct {
	Type       string                  `config:"type"`
	AWSConfig  aws.ConfigAWS           `config:",inline"`
	RuntimeCfg *RuntimeConfig          `config:"runtime_cfg"`
	Fetchers   []*config.C             `config:"fetchers"`
	KubeConfig string                  `config:"kube_config"`
	Period     time.Duration           `config:"period"`
	Processors processors.PluginConfig `config:"processors"`
	BundlePath string                  `config:"bundle_path"`
}

type RuntimeConfig struct {
	ActivatedRules *Benchmarks `config:"activated_rules" yaml:"activated_rules" json:"activated_rules"`
}

type Benchmarks struct {
	CisK8s []string `config:"cis_k8s,omitempty" yaml:"cis_k8s,omitempty" json:"cis_k8s,omitempty"`
	CisEks []string `config:"cis_eks,omitempty" yaml:"cis_eks,omitempty" json:"cis_eks,omitempty"`
}

func New(cfg *config.C) (*Config, error) {
	c, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.Unpack(&c); err != nil {
		return nil, err
	}

	if c.RuntimeCfg != nil && c.RuntimeCfg.ActivatedRules != nil && len(c.RuntimeCfg.ActivatedRules.CisEks) > 0 {
		c.Type = InputTypeEks
	}
	return c, nil
}

func defaultConfig() (*Config, error) {
	ret := &Config{
		Period: 4 * time.Hour,
		Type:   InputTypeVanillaK8s,
	}

	bundle, err := getBundlePath()
	if err != nil {
		return nil, err
	}

	ret.BundlePath = bundle
	return ret, nil
}

func getBundlePath() (string, error) {
	// The bundle resides on the same location as the executable
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(ex), "bundle.tar.gz"), nil
}

// Datastream function to generate the datastream value
func Datastream(namespace string, indexPrefix string) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	return indexPrefix + "-" + namespace
}

type AwsConfigProvider interface {
	InitializeAWSConfig(ctx context.Context, cfg aws.ConfigAWS) (awssdk.Config, error)
}
