package main

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	log.Fatal(initServer())
}

func initServer() error {
	e := echo.New()

	e.HideBanner = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.Recover())
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		return nil
	})

	e.POST("/enqueue", func(c echo.Context) error {
		var input struct {
			Worker          string            `json:"worker"`
			Spider          string            `json:"spider"`
			Args            map[string]string `json:"args"`
			MaxExecDuration string            `json:"max_execution_time"`
			Weight          int               `json:"weight"`
		}

		if err := c.Bind(&input); err != nil {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		maxExecDur, err := time.ParseDuration(input.MaxExecDuration)
		if err != nil {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		queueName, exists := WorkerMethodQueueNameMap[input.Worker]
		if !exists {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		job := NewJob(input.Spider, input.Args, maxExecDur)
		jobAsString, err := job.Serialize()
		if err != nil {
			return c.JSON(500, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		executeNow := false

		if _, exists := c.QueryParams()["force"]; exists {
			executeNow = true
		}

		if !executeNow {
			if input.Worker == "weight" {
				if err := redisConn.ZIncrBy(context.Background(), queueName, float64(input.Weight), jobAsString).Err(); err != nil {
					return c.JSON(500, map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					})
				}
			} else {
				if err := redisConn.LPush(context.Background(), queueName, jobAsString).Err(); err != nil {
					return c.JSON(500, map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					})
				}
			}
		}

		if executeNow {
			result, err := job.Dispatch()
			if err != nil {
				return c.JSON(500, map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				})
			}

			return c.JSON(200, map[string]interface{}{
				"success": true,
				"payload": result,
			})
		}

		return c.JSON(201, map[string]interface{}{
			"success": true,
			"payload": map[string]interface{}{
				"jobid":      job.ID,
				"queue_name": queueName,
				"queue_size": redisConn.ZCard(context.Background(), queueName).Val(),
			},
		})
	})

	return e.Start(*&config.ListenAddr)
}
