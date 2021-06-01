// Package runtime provides the classifier Runtime
package runtime

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var RuntimeConfig *Config

type Config struct {
	ClassifierName string `yaml:"ClassifierName"`
	Sbi            Sbi    `yaml:"sbi"`
}

type Sbi struct {
	RegisterIPv4 string `yaml:"registerIPv4"`
	Port         int    `yaml:"port"`
}

// ParseConf read the yaml file and populate the Config instancce
func ParseConf(file string) error {
	path, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &RuntimeConfig)
	if err != nil {
		return err
	}
	return nil
}

type PDU struct {
	TEID     uint32 `json:"teid" yaml:"teid" bson:"teid"`
	DSCP5    uint8  `json:"dscp_5g" yaml:"dscp_5g" bson:"dscp_5g"`
	DSCPS    uint8  `json:"dscp_satellite" yaml:"dscp_satellite" bson:"dscp_satellite"`
	SliceID  uint8  `json:"slice_id" yaml:"slice_id" bson:"slice_id"`
	Endpoint string `json:"endpoint" yaml:"endpoint" bson:"endpoint"`
	IPv4     string `json:"ipv4" yaml:"ipv4" bson:"ipv4"`
	Ingress  string `json:"ingress" yaml:"ingress" bson:"ingress"`
	IsRAN    bool   `json:"is_ran" yaml:"is_ran" bson:"is_ran"`
}

type ADMControl struct {
	SliceID    uint8  `json:"slice_id" yaml:"slice_id" bson:"slice_id"`
	Throughput int    `json:"throughput" yaml:"throughput" bson:"throughput"`
	Endpoint   string `json:"endpoint" yaml:"endpoint" bson:"endpoint"`
}

type ADM struct {
	Controls []ADMControl `json:"controls" yaml:"controls" bson:"controls"`
	Aware    bool         `json:"slice_aware" yaml:"slice_aware" bson:"slice_aware"`
}
