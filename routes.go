package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	underscore "github.com/ahl5esoft/golang-underscore"
	"github.com/google/uuid"
	"github.com/labstack/echo"
)

func routeHome(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"success":          true,
		"node_name":        nodeName,
		"node_ip":          nodeIP,
		"discovered_nodes": redisConn.HGetAll(redisNodesDiscovery).Val(),
	})
}

func routeDaemonStatus(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"success":   true,
		"status":    "ok",
		"pending":   redisConn.SCard(redisPendingQueueKey).Val(),
		"running":   redisConn.SCard(redisRunningQueueKey).Val(),
		"finished":  redisConn.SCard(redisFinishedQueueKey).Val(),
		"node_name": nodeName,
	})
}

func routeAddVersion(c echo.Context) error {
	var input struct {
		Project string `form:"project" query:"project"`
		Version string `form:"version" query:"version"`
		Egg     string
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	if input.Project == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a project",
		})
	}

	file, err := c.FormFile("egg")
	if err != nil {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	src, err := file.Open()
	if err != nil {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	eggData, err := ioutil.ReadAll(src)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	input.Egg = base64.StdEncoding.EncodeToString(eggData)

	redisConn.SAdd(redisProjectsKey, input.Project)
	redisConn.HSet(redisEggsPrefix+input.Project, input.Version, input.Egg)
	redisConn.LPush(redisVersionsPrefix+input.Project, input.Version)

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"status":  "ok",
		"project": input.Project,
		"version": input.Version,
		"egg":     input.Egg,
		"size":    fmt.Sprintf("%d kb", len(eggData)/1024),
	})
}

func routeListProjects(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"status":   "ok",
		"success":  true,
		"projects": redisConn.SMembers(redisProjectsKey).Val(),
	})
}

func routeListVersions(c echo.Context) error {
	var input struct {
		Project string `json:"project" query:"project" form:"project"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	if input.Project == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a project",
		})
	}

	return c.JSON(200, map[string]interface{}{
		"success":  true,
		"status":   "ok",
		"versions": redisConn.HKeys(redisEggsPrefix + input.Project).Val(),
	})
}

func routeDelProject(c echo.Context) error {
	var input struct {
		Project string `json:"project" query:"project" form:"project"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	if input.Project == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a project",
		})
	}

	redisConn.SRem(input.Project)
	redisConn.Del(redisEggsPrefix + input.Project)
	redisConn.Del(redisVersionsPrefix + input.Project)

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"status":  "ok",
	})
}

func routeDelVersion(c echo.Context) error {
	var input struct {
		Version string `json:"_version" query:"_version" form:"_version"`
		Project string `json:"project" query:"project" form:"project"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	if input.Project == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a project",
		})
	}

	if input.Version == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a version",
		})
	}

	redisConn.HDel(redisEggsPrefix+input.Project, input.Version)
	redisConn.LRem(redisVersionsPrefix+input.Project, 1, input.Version)

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"status":  "ok",
	})
}

func routeListSpiders(c echo.Context) error {
	var input struct {
		Version string `json:"_version" query:"_version" form:"_version"`
		Project string `json:"project" query:"project" form:"project"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	if input.Project == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a project",
		})
	}

	if input.Version == "" {
		input.Version = getProjectLatestVersion(input.Project)
	}

	spiders, err := getProjectSpiders(input.Project, input.Version)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"status":  "ok",
		"spiders": spiders,
	})
}

func routeSchedule(c echo.Context) error {
	project := c.FormValue("project")
	version := c.FormValue("_version")
	spider := c.FormValue("spider")
	jobid := c.FormValue("jobid")

	if project == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a project",
		})
	}

	if spider == "" {
		return c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   "you must specify a spider",
		})
	}

	if version == "" {
		version = getProjectLatestVersion(project)
	}

	if jobid == "" {
		jobid = uuid.New().String()
	}

	query, _ := c.FormParams()

	job := CrawlingJob{
		ID:       jobid,
		Project:  project,
		Spider:   spider,
		Version:  version,
		Settings: query["setting"],
		Args:     []string{},
	}

	delete(query, "project")
	delete(query, "spider")
	delete(query, "_version")
	delete(query, "jobid")
	delete(query, "setting")

	for k, v := range query {
		job.Args = append(job.Args, fmt.Sprintf("%s=%s", k, v[0]))
	}

	job.toPending()

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"status":  "ok",
		"jobid":   jobid,
		"job":     job,
	})
}

func routeListJobs(c echo.Context) error {
	var pending, running, finished []map[string]interface{}

	filter := func(v interface{}, _ int) bool {
		return v != nil
	}

	mapper := func(v interface{}, _ int) map[string]interface{} {
		if v == nil {
			return nil
		}

		var o map[string]interface{}
		json.Unmarshal([]byte(v.(string)), &o)
		return o
	}

	underscore.Chain(redisConn.HMGet(redisJobsKey, redisConn.SMembers(redisPendingQueueKey).Val()...).Val()).Where(filter).Map(mapper).Values().Value(&pending)
	underscore.Chain(redisConn.HMGet(redisJobsKey, redisConn.SMembers(redisRunningQueueKey).Val()...).Val()).Where(filter).Map(mapper).Values().Value(&running)
	underscore.Chain(redisConn.HMGet(redisJobsKey, redisConn.SMembers(redisFinishedQueueKey).Val()...).Val()).Where(filter).Map(mapper).Values().Value(&finished)

	return c.JSON(200, map[string]interface{}{
		"success":  true,
		"status":   "ok",
		"pending":  pending,
		"running":  running,
		"finished": finished,
	})
}

func routeCancelJob(c echo.Context) error {
	jobid := c.FormValue("job")

	if jobid != "" {
		redisConn.SAdd(redisCancelsKey, jobid)
	}

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"status":  "ok",
	})
}

func routeLogView(c echo.Context) error {
	jobid := c.Param("jobid")
	ch := redisConn.Subscribe(redisLogsPrefix + jobid)

	c.Response().Header().Set("Content-Type", "text/plain")
	c.Response().WriteHeader(200)

	readingChannel := ch.Channel()
	closedChannel := c.Response().Writer.(http.CloseNotifier).CloseNotify()
	run := true

	for run && !redisConn.SIsMember(redisFinishedQueueKey, jobid).Val() {
		select {
		case d := <-readingChannel:
			c.Response().Write([]byte(d.Payload))
			c.Response().Flush()
		case <-closedChannel:
			run = false
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	ch.Unsubscribe()
	ch.Close()

	c.Response().Write([]byte("this job may be ended"))

	return nil
}
