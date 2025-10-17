package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
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

	// Last created file directory
	LastCreatedDir string
}

// load environment variables from .env or system environment.
func loadEnv() (JobConfig, error) {
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	godotenv.Load(".env")

	apiUrl := os.Getenv("IMMICH_API_URL")

	// Try to read API key from Docker secret first, then fallback to environment variable
	apiKey := readSecretOrEnv("IMMICH_API_KEY", "/run/secrets/immich_api_key")

	if apiUrl == "" || apiKey == "" {
		return JobConfig{}, fmt.Errorf("IMMICH_API_URL or IMMICH_API_KEY environment variable/secret is not set")
	}

	assetTimezoneFixerCronExprStr := os.Getenv("CRON_EXPRESSION")
	assetTimezoneFixerCronExpr, err := cronParser.Parse(assetTimezoneFixerCronExprStr)
	if err != nil {
		return JobConfig{}, fmt.Errorf("CRON_EXPRESSION environment variable is not set")
	}

	lastCreatedDir := "./"
	if lcd := os.Getenv("LAST_CREATED_DIR"); lcd != "" {
		lastCreatedDir = lcd
	}

	return JobConfig{
		ApiUrl:         apiUrl,
		ApiKey:         apiKey,
		CronExpression: assetTimezoneFixerCronExpr,
		LastCreatedDir: lastCreatedDir,
	}, nil
}

// readSecretOrEnv reads a value from Docker secret file first, then falls back to environment variable
func readSecretOrEnv(envName, secretPath string) string {
	// Try to read from Docker secret file first
	if secretData, err := os.ReadFile(secretPath); err == nil {
		return strings.TrimSpace(string(secretData))
	}

	// Fallback to environment variable
	return os.Getenv(envName)
}
