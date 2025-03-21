package main

import (
	"cdn/api/router"
	"fmt"
	"runtime"

	"github.com/robfig/cron/v3"
)

func main() {
    // Initialize the cron scheduler
    c := cron.New()

    router.CleanAllFilesFromFolder()

    fmt.Println("Starting cron job")

    // Schedule the cron job (runs every minute)
    // Run cleanup every second
    _, err := c.AddFunc("@every 1s", func() {
        fmt.Println("Running files listing check...")
        router.ListAllFilesFromFolder()
    })

    if err != nil {
		fmt.Println("Error adding cron job:", err)
		return
	}
    
    // Add memory cleanup job
    c.AddFunc("@every 5m", func() {
        fmt.Println("Running garbage collection")
        runtime.GC()
    })

    // Start the cron scheduler
    c.Start()
    
    // Start the router
    router.InitRoutes()

    select {}

}