package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type Client struct {
	client     *azcosmos.Client
	database   *azcosmos.DatabaseClient
	containers map[string]*azcosmos.ContainerClient
}

type Config struct {
	Endpoint string
	Key      string
	Database string
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	cred, err := azcosmos.NewKeyCredential(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	client, err := azcosmos.NewClientWithKey(cfg.Endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	database, err := client.NewDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	c := &Client{
		client:     client,
		database:   database,
		containers: make(map[string]*azcosmos.ContainerClient),
	}

	containers := []string{"organizations", "applications", "service_keys", "api_keys", "users"}
	for _, containerName := range containers {
		container, err := database.NewContainer(containerName)
		if err != nil {
			return nil, fmt.Errorf("failed to get container %s: %w", containerName, err)
		}
		c.containers[containerName] = container
	}

	return c, nil
}

// GetContainer returns a container by name
func (c *Client) GetContainer(name string) (*azcosmos.ContainerClient, error) {
	container, ok := c.containers[name]
	if !ok {
		return nil, fmt.Errorf("container %s not found", name)
	}
	return container, nil
}

// GenerateID generates a random ID
func GenerateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// handleCosmosError converts Cosmos DB errors to more readable errors
func handleCosmosError(err error) error {
	if err == nil {
		return nil
	}

	// log error
	fmt.Printf("Cosmos error: %v\n", err)

	var respErr *azcore.ResponseError
	if err, ok := err.(*azcore.ResponseError); ok {
		respErr = err
	}

	if respErr != nil {
		switch respErr.StatusCode {
		case 404:
			return fmt.Errorf("not found")
		case 409:
			return fmt.Errorf("conflict: resource already exists")
		case 400:
			return fmt.Errorf("bad request: %s", respErr.Error())
		default:
			return fmt.Errorf("cosmos db error: %s", respErr.Error())
		}
	}

	return fmt.Errorf("cosmos db error: %w", err)
}
