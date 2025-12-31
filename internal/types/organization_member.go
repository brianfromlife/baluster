package types

import "time"

// OrganizationMember represents a membership relationship between a user and an organization
type OrganizationMember struct {
	ID             string    `json:"id" cosmosdb:"id"`
	PartitionKey   string    `json:"-" cosmosdb:"_partitionKey"`
	EntityType     string    `json:"entity_type"` // "organization_member" discriminator
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// GetPartitionKey returns the partition key for Cosmos DB
func (om *OrganizationMember) GetPartitionKey() string {
	return om.OrganizationID
}

// GetID returns the document ID for Cosmos DB
func (om *OrganizationMember) GetID() string {
	return om.OrganizationID + "_" + om.UserID
}

