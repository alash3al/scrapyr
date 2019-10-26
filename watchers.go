package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archiver"
)

func runWatchers() {
	go (func() {
		ticker := time.NewTicker(time.Duration(*flagSyncInterval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			syncEggs()
		}
	})()

	go terminator()

	go execWorkers()

	go cleanJobs()
}

func syncEggs() error {
	if err := os.MkdirAll(*flagCacheDir, 0755); err != os.ErrExist && err != nil {
		return err
	}

	for _, project := range redisConn.SMembers(redisProjectsKey).Val() {
		projectPath := filepath.Join(*flagCacheDir, project)

		if err := os.MkdirAll(filepath.Join(projectPath, "eggs"), 0755); err != os.ErrExist && err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Join(projectPath, "src"), 0755); err != os.ErrExist && err != nil {
			return err
		}

		for version, egg := range redisConn.HGetAll(redisEggsPrefix + project).Val() {
			filename := filepath.Join(projectPath, "eggs", version) + ".egg"
			if _, err := os.Stat(filename); err == nil {
				continue
			}

			eggFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}

			bindata, err := base64.StdEncoding.DecodeString(egg)
			if err != nil {
				eggFile.Close()
				return err
			}

			if _, err := eggFile.Write(bindata); err != nil {
				eggFile.Close()
				return err
			}
			eggFile.Close()

			if err := archiver.NewZip().Unarchive(filename, filepath.Join(projectPath, "src", version)); err != nil {
				return err
			}

			data := fmt.Sprintf("[settings]\ndefault = %s.settings", project)
			if err := ioutil.WriteFile(filepath.Join(projectPath, "src", version, "scrapy.cfg"), []byte(data), 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

func terminator() {
	for {
		id := redisConn.SPop(redisCancelsKey).Val()
		if id == "" {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if cmd, ok := cmds[id]; ok {
			cmd.Process.Kill()
		} else {
			redisConn.HDel(redisJobsKey, id)
			redisConn.SRem(redisPendingQueueKey, id)
		}
	}
}

func execWorkers() {
	for {
		if redisConn.SCard(redisPendingQueueKey).Val() > 0 {
			ids := redisConn.SPopN(redisPendingQueueKey, int64(*flagMaxWorkers)).Val()
			for _, id := range ids {
				job := newJobByID(id)

				go workersPool.Process(job)
			}
		}

		time.Sleep(time.Duration(*flagPollInterval) * time.Millisecond)
	}
}

func cleanJobs() {
	for {
		time.Sleep(5 * time.Second)

		if redisConn.SCard(redisFinishedQueueKey).Val() < *flagMaxToKeep {
			continue
		}

		toBeDeleted := redisConn.SPopN(redisFinishedQueueKey, *flagMaxToKeep).Val()
		redisConn.HDel(redisJobsKey, toBeDeleted...)
	}
}
