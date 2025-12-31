package types

import "time"

// ApiKey represents token that are used to authenticate requests to Baluster itself
type ApiKey struct {
	ID                string     `json:"id" cosmosdb:"id"`
	PartitionKey      string     `json:"-" cosmosdb:"_partitionKey"`
	EntityType        string     `json:"entity_type"` // "api_key" discriminator
	OrganizationID    string     `json:"organization_id"`
	ApplicationID     string     `json:"application_id"`
	Name              string     `json:"name"`
	TokenValue        string     `json:"token_value"`
	ExpiresAt         *time.Time `json:"expires_at"`
	CreatedByUserID   string     `json:"created_by_user_id"`
	CreatedByGitHubID string     `json:"created_by_github_id"`
	CreatedByUsername string     `json:"created_by_username"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (t *ApiKey) GetPartitionKey() string {
	return t.OrganizationID
}

func (t *ApiKey) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return t.ExpiresAt.Before(time.Now())
}
