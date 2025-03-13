package main

import (
	"cdn/api/router"
	"fmt"
	"log"

	"github.com/robfig/cron"
)


func main() {
// Initialize the cron scheduler

	router.CleanAllFilesFromFolder()

	fmt.Println("Starting cron job")
	c := cron.New()

	// Schedule the cron job (runs every minute)
	err := c.AddFunc("@every 1s", router.ListAllFilesFromFolder)
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	// Start the cron scheduler
	c.Start()
	defer c.Stop() // Ensure the cron job stops when the app exits


	router.InitRoutes()
}