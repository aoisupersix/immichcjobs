package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	config, err := loadEnv()
	if err != nil {
		slog.Error("Failed to load environment variables", "error", err)
		return
	}

	c := cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))
	c.Schedule(config.CronExpression, cron.FuncJob(func() { assetTimezoneFixer(config) }))

	fmt.Println("Starting Immich Custom Jobs...")
	fmt.Printf("API URL: %s\n", config.ApiUrl)
	fmt.Printf("Job Next Run: %s\n", config.CronExpression.Next(time.Now()).Format(time.RFC3339))

	c.Start()
	defer c.Stop()

	// Keep the application running
	select {}
}

// Custom job configuration.
type JobConfig struct {
	// Immich API URL
	ApiUrl string

	// Immich API Key
	ApiKey string

	// Cron expression for the job
	CronExpression cron.Schedule
}

// load environment variables from .env or system environment.
func loadEnv() (JobConfig, error) {
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	godotenv.Load(".env")

	apiUrl := os.Getenv("IMMICH_API_URL")
	apiKey := os.Getenv("IMMICH_API_KEY")
	if apiUrl == "" || apiKey == "" {
		return JobConfig{}, fmt.Errorf("IMMICH_API_URL or IMMICH_API_KEY environment variable is not set")
	}

	assetTimezoneFixerCronExprStr := os.Getenv("CRON_EXPRESSION")
	assetTimezoneFixerCronExpr, err := cronParser.Parse(assetTimezoneFixerCronExprStr)
	if err != nil {
		return JobConfig{}, fmt.Errorf("CRON_EXPRESSION environment variable is not set")
	}

	return JobConfig{
		ApiUrl:         apiUrl,
		ApiKey:         apiKey,
		CronExpression: assetTimezoneFixerCronExpr,
	}, nil
}
