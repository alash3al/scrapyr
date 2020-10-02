package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/rs/xid"
)

type Job struct {
	ID          string            `json:"id"`
	Spider      string            `json:"spider"`
	Args        map[string]string `json:"args"`
	MaxExecTime time.Duration     `json:"max_execution_time"`
}

func NewJob(spider string, args map[string]string, timeout time.Duration) *Job {
	return &Job{
		ID:          fmt.Sprintf("job-%s-%d", xid.New().String(), time.Now().Unix()),
		Spider:      spider,
		Args:        args,
		MaxExecTime: timeout,
	}
}

func UnserializeJob(s string) (*Job, error) {
	var j Job

	if err := json.Unmarshal([]byte(s), &j); err != nil {
		return nil, err
	}

	return &j, nil
}

func (j Job) Serialize() (string, error) {
	data, err := json.Marshal(j)

	if data != nil {
		return string(data), err
	}

	return "", err
}

func (j Job) Dispatch() (interface{}, error) {
	if j.ID == "" {
		return nil, errors.New("unable to generate a job id")
	}

	jobfilename := filepath.Join(config.Scrapy.ItemsDir, j.ID+".json")

	j.Args["jobid"] = j.ID
	j.Args["jobfilename"] = jobfilename

	argv := []string{config.Scrapy.PythonBin, "-m", "scrapy", "crawl", "-o", jobfilename, "-L", "ERROR", j.Spider}

	for k, v := range j.Args {
		argv = append(argv, "-a", fmt.Sprintf(`%s=%s`, k, v))
	}

	if err := cmd(j.MaxExecTime, argv, nil); err != nil {
		return nil, err
	}

	jobdata, err := ioutil.ReadFile(jobfilename)
	if err != nil {
		return nil, err
	}

	var jobResult interface{}

	if err := json.Unmarshal(jobdata, &jobResult); err != nil {
		return nil, err
	}

	return jobResult, nil
}
