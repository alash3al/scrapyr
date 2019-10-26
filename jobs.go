package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Jeffail/tunny"
)

var (
	workersPool *tunny.Pool
)

func init() {
	workersPool = tunny.NewFunc(*flagMaxWorkers, func(j interface{}) interface{} {
		job := j.(CrawlingJob)

		return job.Exec()
	})
}

// CrawlingJob a scrapy crawling job
type CrawlingJob struct {
	raw       string
	ID        string    `json:"id"`
	Project   string    `json:"project"`
	Spider    string    `json:"spider"`
	Version   string    `json:"version"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Elapsed   float64   `json:"elapsed,omitempty"`
	Settings  []string  `json:"settings"`
	Args      []string  `json:"args"`
}

func newJobByID(id string) CrawlingJob {
	var j CrawlingJob

	json.Unmarshal([]byte(redisConn.HGet(redisJobsKey, id).Val()), &j)

	return j
}

func (j CrawlingJob) cmd() *exec.Cmd {
	args := []string{"-m", "scrapy", "crawl", j.Spider}

	for _, s := range j.Settings {
		args = append(args, "-s", s)
	}

	for _, a := range j.Args {
		args = append(args, "-a", a)
	}

	out := NewRedisWriter(redisLogsPrefix + j.ID)

	cmd := exec.Command("python", args...)
	cmd.Dir = filepath.Join(*flagCacheDir, j.Project, "src", j.Version)
	cmd.Stdout = out
	cmd.Stderr = out

	return cmd
}

// Exec executes the job
func (j CrawlingJob) Exec() error {
	cmd := j.cmd()

	if err := cmd.Start(); err != nil {
		return err
	}

	cmds[j.ID] = cmd

	j.toRunning()

	err := cmd.Wait()

	j.toFinished()

	return err
}

func (j CrawlingJob) String() string {
	jsn, _ := json.Marshal(j)

	return string(jsn)
}

func (j CrawlingJob) toPending() {
	redisConn.SRem(redisFinishedQueueKey, j.ID)
	redisConn.SAdd(redisPendingQueueKey, j.ID)
	redisConn.HSet(redisJobsKey, j.ID, j.String())
}

func (j CrawlingJob) toRunning() {
	j.StartTime = time.Now()

	redisConn.SRem(redisPendingQueueKey, j.ID)
	redisConn.SAdd(redisRunningQueueKey, j.ID)
	redisConn.HSet(redisJobsKey, j.ID, j.String())
}

func (j CrawlingJob) toFinished() {
	j.EndTime = time.Now()
	j.Elapsed = j.EndTime.Sub(j.StartTime).Seconds()

	redisConn.SRem(redisRunningQueueKey, j.ID)
	redisConn.SAdd(redisFinishedQueueKey, j.ID)
	redisConn.HSet(redisJobsKey, j.ID, j.String())
}
