package types

import "time"

// AuditAction represents the type of action performed on an entity
type AuditAction string

const (
	AuditActionCreated AuditAction = "created"
	AuditActionUpdated AuditAction = "updated"
	AuditActionDeleted AuditAction = "deleted"
)

// AuditHistory represents an audit history record for tracking entity changes
// These records are stored in the same container as their respective entity
// with a discriminator field (EntityType) to differentiate them
type AuditHistory struct {
	ID                string      `json:"id" cosmosdb:"id"`
	PartitionKey      string      `json:"-" cosmosdb:"_partitionKey"`
	EntityType        string      `json:"entity_type"` // "audit_history" discriminator
	EntityID          string      `json:"entity_id"`   // ID of the entity being audited
	OrganizationID    string      `json:"organization_id"`
	Action            AuditAction `json:"action"` // created, updated, deleted
	CreatedByUserID   string      `json:"created_by_user_id"`
	CreatedByGitHubID string      `json:"created_by_github_id"`
	CreatedByUsername string      `json:"created_by_username"`
	CreatedAt         time.Time   `json:"created_at"`
}

// GetPartitionKey returns the partition key for Cosmos DB (organization_id)
func (a *AuditHistory) GetPartitionKey() string {
	return a.OrganizationID
}
