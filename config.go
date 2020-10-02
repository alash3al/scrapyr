package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl"
)

type Config struct {
	ListenAddr string `hcl:"listen_addr"`

	RedisDSN string `hcl:"redis_dsn"`

	Scrapy struct {
		ProjectDir string `hcl:"project_dir"`
		PythonBin  string `hcl:"python_bin"`
		ItemsDir   string `hcl:"items_dir"`
	} `hcl:"scrapy"`

	Workers map[string]WorkerConfig `hcl:"worker"`
}

type WorkerConfig struct {
	Method   WorkerMethod `hcl:"use"`
	MaxProcs int          `hcl:"max_procs"`
}

type WorkerMethod string

const (
	WorkerMethodFIFO   WorkerMethod = "fifo"
	WorkerMethodLIFO   WorkerMethod = "lifo"
	WorkerMethodWeight WorkerMethod = "weight"
)

var (
	WorkerMethodQueueNameMap = map[string]string{
		"fifo":   fifoQueueName,
		"lifo":   lifoQueueName,
		"weight": weightQueueName,
	}
)

func ParseConfigFile(filename string) (*Config, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fileContent = []byte(os.ExpandEnv(string(fileContent)))

	var config Config

	if err := hcl.Unmarshal(fileContent, &config); err != nil {
		return nil, err
	}

	// TODO: run some config tests
	{
		for _, dir := range []string{config.Scrapy.ProjectDir, config.Scrapy.ItemsDir} {
			dirStat, err := os.Stat(dir)
			if err != nil {
				return nil, err
			}

			if !dirStat.IsDir() {
				return nil, fmt.Errorf("%s isn't a directory", dir)
			}
		}

		if _, err := os.Stat(config.Scrapy.PythonBin); err != nil {
			return nil, err
		}
	}

	return &config, nil
}
