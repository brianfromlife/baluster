package types

import "time"

// Organization represents a top-level entity
type Organization struct {
	ID             string    `json:"id" cosmosdb:"id"`
	PartitionKey   string    `json:"-" cosmosdb:"_partitionKey"`
	EntityType     string    `json:"entity_type"` // "organization" discriminator
	OrganizationID string    `json:"organization_id"` // Same as ID, used as partition key
	Name           string    `json:"name"`
	MemberIDs      []string  `json:"member_ids"` // Deprecated: kept for backward compatibility during migration
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetPartitionKey returns the partition key for Cosmos DB
// Uses organization_id (which equals ID) to allow sharing partition with organization_members
func (o *Organization) GetPartitionKey() string {
	// If OrganizationID is set, use it; otherwise fall back to ID
	if o.OrganizationID != "" {
		return o.OrganizationID
	}
	return o.ID
}
