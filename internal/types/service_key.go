package types

import (
	"slices"
	"time"
)

// ApplicationAccess defines which permissions a service key has for an application
type ApplicationAccess struct {
	ApplicationID   string   `json:"application_id"`
	ApplicationName string   `json:"application_name"`
	Permissions     []string `json:"permissions"`
}

// ServiceKey represents a token that grants access to multiple applications with specific permissions
type ServiceKey struct {
	ID                string              `json:"id" cosmosdb:"id"`
	PartitionKey      string              `json:"-" cosmosdb:"_partitionKey"`
	EntityType        string              `json:"entity_type"` // "service_key" discriminator
	OrganizationID    string              `json:"organization_id"`
	Name              string              `json:"name"`
	TokenValue        string              `json:"token_value"`
	Applications      []ApplicationAccess `json:"applications"`
	ExpiresAt         *time.Time          `json:"expires_at"`
	CreatedByUserID   string              `json:"created_by_user_id"`
	CreatedByGitHubID string              `json:"created_by_github_id"`
	CreatedByUsername string              `json:"created_by_username"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

func (t *ServiceKey) GetPartitionKey() string {
	return t.OrganizationID
}

func (t *ServiceKey) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return t.ExpiresAt.Before(time.Now())
}

// HasAccessToApplication checks if the service key has access to a specific application
func (t *ServiceKey) HasAccessToApplication(applicationName string) *ApplicationAccess {
	for _, app := range t.Applications {
		if app.ApplicationName == applicationName {
			return &app
		}
	}
	return nil
}

// HasPermission checks if the service key has a specific permission for an application
func (t *ServiceKey) HasPermission(applicationName string, permission string) bool {
	access := t.HasAccessToApplication(applicationName)
	if access == nil {
		return false
	}
	return slices.Contains(access.Permissions, permission)
}
