package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/brianfromlife/baluster/internal/types"
)

// OrganizationRepository handles organization storage operations
type OrganizationRepository struct {
	client    *Client
	container *azcosmos.ContainerClient
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(client *Client) (*OrganizationRepository, error) {
	container, err := client.GetContainer("organizations")
	if err != nil {
		return nil, err
	}
	return &OrganizationRepository{
		client:    client,
		container: container,
	}, nil
}

// Create creates a new organization
func (r *OrganizationRepository) Create(ctx context.Context, org *types.Organization) error {
	// Set entity_type and ensure organization_id matches id
	org.EntityType = "organization"
	if org.OrganizationID == "" {
		org.OrganizationID = org.ID
	}
	org.PartitionKey = org.GetPartitionKey()
	item, err := json.Marshal(org)
	if err != nil {
		return fmt.Errorf("failed to marshal organization: %w", err)
	}

	_, err = r.container.CreateItem(ctx, azcosmos.NewPartitionKeyString(org.PartitionKey), item, nil)
	return handleCosmosError(err)
}

// Delete deletes an organization by ID
func (r *OrganizationRepository) Delete(ctx context.Context, id string) error {
	// Partition key is now organization_id (which equals id for organizations)
	_, err := r.container.DeleteItem(ctx, azcosmos.NewPartitionKeyString(id), id, nil)
	return handleCosmosError(err)
}

// Get retrieves an organization by ID
func (r *OrganizationRepository) Get(ctx context.Context, id string) (*types.Organization, error) {
	// Partition key is now organization_id (which equals id for organizations)
	itemResponse, err := r.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(id), id, nil)
	if err != nil {
		return nil, handleCosmosError(err)
	}

	var org types.Organization
	if err := json.Unmarshal(itemResponse.Value, &org); err != nil {
		return nil, fmt.Errorf("failed to unmarshal organization: %w", err)
	}

	// Verify entity type to ensure we got an organization, not a member
	if org.EntityType != "" && org.EntityType != "organization" {
		return nil, fmt.Errorf("organization not found")
	}

	return &org, nil
}

// List lists all organizations
func (r *OrganizationRepository) List(ctx context.Context) ([]*types.Organization, error) {
	// Filter by entity_type to only get organizations, not members
	query := "SELECT * FROM c WHERE (c.entity_type = 'organization' OR NOT IS_DEFINED(c.entity_type))"
	// Use NewPartitionKey() for cross-partition query (empty partition key list)
	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKey(), nil)

	var orgs []*types.Organization
	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range queryResponse.Items {
			var org types.Organization
			if err := json.Unmarshal(item, &org); err != nil {
				return nil, fmt.Errorf("failed to unmarshal organization: %w", err)
			}
			// Skip if it's actually a member (backward compatibility check)
			if org.EntityType == "organization_member" {
				continue
			}
			orgs = append(orgs, &org)
		}
	}

	return orgs, nil
}

// ListByMemberID lists organizations where the user is a member
func (r *OrganizationRepository) ListByMemberID(ctx context.Context, userID string) ([]*types.Organization, error) {
	var orgs []*types.Organization
	orgMap := make(map[string]*types.Organization) // Use map to deduplicate

	// Step 1: Query for organization_member records where user_id matches (new system)
	memberQuery := fmt.Sprintf("SELECT * FROM c WHERE c.entity_type = 'organization_member' AND c.user_id = '%s'", userID)
	memberPager := r.container.NewQueryItemsPager(memberQuery, azcosmos.NewPartitionKey(), nil)

	for memberPager.More() {
		memberResponse, err := memberPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range memberResponse.Items {
			var member types.OrganizationMember
			if err := json.Unmarshal(item, &member); err != nil {
				// Skip if unmarshal fails - might be an organization record
				continue
			}
			if member.EntityType != "organization_member" {
				continue
			}

			// Fetch the organization for this member record
			org, err := r.Get(ctx, member.OrganizationID)
			if err != nil {
				// If org not found, skip this member
				continue
			}
			// Add to map to avoid duplicates
			orgMap[org.ID] = org
		}
	}

	// Step 2: Query organizations where member_ids array contains the user ID (backward compatibility)
	legacyQuery := fmt.Sprintf("SELECT * FROM c WHERE ((c.entity_type = 'organization' OR NOT IS_DEFINED(c.entity_type)) AND ARRAY_CONTAINS(c.member_ids, '%s'))", userID)
	legacyPager := r.container.NewQueryItemsPager(legacyQuery, azcosmos.NewPartitionKey(), nil)

	for legacyPager.More() {
		legacyResponse, err := legacyPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range legacyResponse.Items {
			var org types.Organization
			if err := json.Unmarshal(item, &org); err != nil {
				return nil, fmt.Errorf("failed to unmarshal organization: %w", err)
			}
			// Skip if it's actually a member record
			if org.EntityType == "organization_member" {
				continue
			}
			// Add to map to avoid duplicates
			orgMap[org.ID] = &org
		}
	}

	// Convert map to slice
	for _, org := range orgMap {
		orgs = append(orgs, org)
	}

	return orgs, nil
}
