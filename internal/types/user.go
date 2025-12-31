package types

import "time"

// User represents an authenticated user
type User struct {
	ID           string    `json:"id" cosmosdb:"id"`
	PartitionKey string    `json:"-" cosmosdb:"_partitionKey"`
	GitHubID     string    `json:"github_id"`
	Username     string    `json:"username"`
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetPartitionKey returns the partition key for Cosmos DB
func (u *User) GetPartitionKey() string {
	return u.GitHubID
}
