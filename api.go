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
    // Run cleanup every second
    c.AddFunc("@every 1s", router.ListAllFilesFromFolder)

    // Add a channel to keep the program running
    stop := make(chan struct{})
    go func() {
        <-stop // This will block forever
    }()
    
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