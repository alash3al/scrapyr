package main

import (
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.GET("/", routeHome)

	e.GET("daemonstatus.json", routeDaemonStatus)
	e.POST("addversion.json", routeAddVersion)
	e.GET("listprojects.json", routeListProjects)
	e.GET("listversions.json", routeListVersions)
	e.GET("listspiders.json", routeListSpiders)
	e.POST("delproject.json", routeDelProject)
	e.POST("delversion.json", routeDelVersion)
	e.POST("schedule.json", routeSchedule)
	e.GET("listjobs.json", routeListJobs)
	e.POST("cancel.json", routeCancelJob)
	e.GET("logs/:jobid", routeLogView)

	log.Fatal(e.Start(*flagListenAddr))
}
