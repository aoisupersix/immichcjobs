package immichclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/types"
)

// NewClientWithKey creates a new Immich client with the provided API key.
func NewClientWithKey(apiUrl string, apiKey string) (*ClientWithResponses, error) {
	client, err := NewClientWithResponses(apiUrl, WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-api-key", apiKey)
		return nil
	}))

	return client, err
}

type FindAllAssetsOpts struct {
	LastCreated *time.Time
	LibraryId   *types.UUID
	DeviceId    *string
}

type FindAllAssetsOption func(*FindAllAssetsOpts)

func WithLastCreated(t *time.Time) FindAllAssetsOption {
	return func(o *FindAllAssetsOpts) {
		o.LastCreated = t
	}
}

func WithLibraryId(id *types.UUID) FindAllAssetsOption {
	return func(o *FindAllAssetsOpts) {
		o.LibraryId = id
	}
}

func WithDeviceId(id *string) FindAllAssetsOption {
	return func(o *FindAllAssetsOpts) {
		o.DeviceId = id
	}
}

// FindAllAssets retrieves all assets created after the specified time.
func FindAllAssets(client *ClientWithResponses, ctx context.Context, opts ...FindAllAssetsOption) ([]AssetResponseDto, error) {
	options := &FindAllAssetsOpts{
		LastCreated: nil,
		LibraryId:   nil,
		DeviceId:    nil,
	}
	for _, opt := range opts {
		opt(options)
	}

	var assets []AssetResponseDto
	for page := 1; ; page++ {
		pageFloat := float32(page)
		resp, err := client.SearchAssetsWithResponse(ctx, SearchAssetsJSONRequestBody{
			CreatedAfter: options.LastCreated,
			LibraryId:    options.LibraryId,
			DeviceId:     options.DeviceId,
			Order:        &[]AssetOrder{"asc"}[0],
			Page:         &pageFloat,
			Size:         &[]float32{200}[0],
			WithExif:     &[]bool{true}[0]})
		if err != nil {
			return nil, fmt.Errorf("failed to create search assets request: %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("search assets request failed with status: %s", resp.Status())
		}
		if len(resp.JSON200.Assets.Items) == 0 {
			return assets, nil
		}

		assets = append(assets, resp.JSON200.Assets.Items...)
	}
}
