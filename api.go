package main

import (
	"cdn/api/router"
	"fmt"
	"runtime"

	"github.com/robfig/cron"
)

func main() {
    // Initialize the cron scheduler
    c := cron.New()

    router.CleanAllFilesFromFolder()

    fmt.Println("Starting cron job")

    // Schedule the cron job (runs every minute)
    c.AddFunc("@every 1m", router.CleanAllFilesFromFolder)
    
    // Add memory cleanup job
    c.AddFunc("@every 5m", func() {
        fmt.Println("Running garbage collection")
        runtime.GC()
    })

    // Start the cron scheduler
    c.Start()
    
    // Start the router
    router.InitRoutes()
}