package main

import (
	"context"
	"immich-custom-jobs/immichclient"
	"immich-custom-jobs/jobstate"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	JOBNAME = "asset_timezone_fixer"
)

func assetTimezoneFixer(conf JobConfig) {
	slog.Info("Running asset timezone fixer job")

	jstLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("Failed to load JST location", "error", err)
		return
	}

	lastCreated, err := jobstate.ReadLastCreated(JOBNAME, conf.LastCreatedDir)
	if err != nil {
		slog.Error("Failed to read last created timestamp", "error", err)
		return
	}

	ctx := context.Background()
	client, err := immichclient.NewClientWithKey(conf.ApiUrl, conf.ApiKey)
	if err != nil {
		slog.Error("Failed to create Immich client", "error", err)
		return
	}
	assets, err := immichclient.FindAllAssets(client, ctx, immichclient.WithLastCreated(lastCreated))
	if err != nil {
		slog.Error("Failed to fetch assets", "error", err)
		return
	}

	for _, asset := range assets {
		assetId, err := uuid.Parse(asset.Id)
		if err != nil {
			slog.Error("Failed to parse asset ID", "id", asset.Id, "error", err)
			return
		}
		slog.Info("Processing asset", "id", asset.Id, "createdAt", asset.CreatedAt)
		if asset.ExifInfo.TimeZone == nil || strings.ToUpper(*asset.ExifInfo.TimeZone) != "UTC" {
			slog.Info("Skipping asset because timezone is not UTC", "id", asset.Id, "timezone", asset.ExifInfo.TimeZone)
			err = jobstate.WriteLastCreated(JOBNAME, conf.LastCreatedDir, &asset.CreatedAt)
			if err != nil {
				slog.Error("Failed to write last created timestamp", "error", err)
				return
			}
			continue
		}
		jstTime := asset.LocalDateTime.In(jstLocation)
		jstTimeStr := jstTime.Format("2006-01-02 15:04:05-07:00")
		slog.Info("Updating asset timezone", "id", asset.Id, "newLocalDateTime", jstTime, "newTimeZone", "Asia/Tokyo")
		resp, err := client.UpdateAsset(ctx, assetId, immichclient.UpdateAssetJSONRequestBody{
			DateTimeOriginal: &jstTimeStr,
		})
		if err != nil {
			slog.Error("Failed to update asset", "error", err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			slog.Error("Failed to update asset", "status", resp.Status, "statusCode", resp.StatusCode)
			return
		}

		err = jobstate.WriteLastCreated(JOBNAME, conf.LastCreatedDir, &asset.CreatedAt)
		if err != nil {
			slog.Error("Failed to write last created timestamp", "error", err)
			return
		}
	}
}
