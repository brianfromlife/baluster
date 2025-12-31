package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/brianfromlife/baluster/internal/types"
)

// UserRepository handles user storage operations
type UserRepository struct {
	client    *Client
	container *azcosmos.ContainerClient
}

// NewUserRepository creates a new user repository
func NewUserRepository(client *Client) (*UserRepository, error) {
	container, err := client.GetContainer("users")
	if err != nil {
		return nil, err
	}
	return &UserRepository{
		client:    client,
		container: container,
	}, nil
}

// CreateOrUpdate creates or updates a user
func (r *UserRepository) CreateOrUpdate(ctx context.Context, user *types.User) error {
	if user.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	if user.GitHubID == "" {
		return fmt.Errorf("user GitHubID is required")
	}

	user.PartitionKey = user.GetPartitionKey()
	item, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = r.container.UpsertItem(ctx, azcosmos.NewPartitionKeyString(user.PartitionKey), item, nil)
	if err != nil {
		return fmt.Errorf("failed to upsert user (id: %s, github_id: %s, partition_key: %s): %w", user.ID, user.GitHubID, user.PartitionKey, handleCosmosError(err))
	}
	return nil
}

// GetByGitHubID retrieves a user by GitHub ID
func (r *UserRepository) GetByGitHubID(ctx context.Context, githubID string) (*types.User, error) {
	// Query by github_id field since that's the partition key
	// The item ID is user.ID, not githubID, so we need to query instead of ReadItem
	query := fmt.Sprintf("SELECT * FROM c WHERE c.github_id = '%s'", githubID)
	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(githubID), nil)

	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range queryResponse.Items {
			var user types.User
			if err := json.Unmarshal(item, &user); err != nil {
				return nil, fmt.Errorf("failed to unmarshal user: %w", err)
			}
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// GetByID retrieves a user by ID using ReadItem
// id is the Cosmos DB item ID, githubID is the partition key
func (r *UserRepository) GetByID(ctx context.Context, id string, githubID string) (*types.User, error) {
	itemResponse, err := r.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(githubID), id, nil)
	if err != nil {
		return nil, handleCosmosError(err)
	}

	var user types.User
	if err := json.Unmarshal(itemResponse.Value, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}
