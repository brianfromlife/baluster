package types

import "time"

// Application belongs to an organization and defines available permissions
type Application struct {
	ID                string    `json:"id" cosmosdb:"id"`
	PartitionKey      string    `json:"-" cosmosdb:"_partitionKey"`
	EntityType        string    `json:"entity_type"` // "application" discriminator
	OrganizationID    string    `json:"organization_id"`
	Name              string    `json:"name"` // use underscores instead of spaces (e.g., user_service)
	Description       string    `json:"description"`
	Permissions       []string  `json:"permissions"`
	CreatedByUserID   string    `json:"created_by_user_id"`
	CreatedByGitHubID string    `json:"created_by_github_id"`
	CreatedByUsername string    `json:"created_by_username"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// GetPartitionKey returns the partition key for Cosmos DB
func (a *Application) GetPartitionKey() string {
	return a.OrganizationID
}
