package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/brianfromlife/baluster/internal/types"
)

// OrganizationMemberRepository handles organization membership storage operations
type OrganizationMemberRepository struct {
	client    *Client
	container *azcosmos.ContainerClient
	orgRepo   *OrganizationRepository // For backward compatibility with MemberIDs array
}

// NewOrganizationMemberRepository creates a new organization member repository
func NewOrganizationMemberRepository(client *Client) (*OrganizationMemberRepository, error) {
	// Use the same container as organizations (they share the partition)
	container, err := client.GetContainer("organizations")
	if err != nil {
		return nil, err
	}
	orgRepo, err := NewOrganizationRepository(client)
	if err != nil {
		return nil, err
	}
	return &OrganizationMemberRepository{
		client:    client,
		container: container,
		orgRepo:   orgRepo,
	}, nil
}

// IsMember checks if a user is a member of an organization
// First checks organization_member records, then falls back to the old MemberIDs array for backward compatibility
func (r *OrganizationMemberRepository) IsMember(ctx context.Context, orgID, userID string) (bool, error) {
	memberID := orgID + "_" + userID
	partitionKey := orgID

	// Check for organization_member record in the organizations container
	_, err := r.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), memberID, nil)
	if err != nil {
		// Check if it's a 404 error (member not found)
		var respErr *azcore.ResponseError
		if err, ok := err.(*azcore.ResponseError); ok {
			respErr = err
		}
		if respErr != nil && respErr.StatusCode == 404 {
			// Member not found, check old MemberIDs array for backward compatibility
			org, getErr := r.orgRepo.Get(ctx, orgID)
			if getErr != nil {
				// If we can't get the org, return false (not a member)
				return false, nil
			}
			// Check if userID is in the MemberIDs array
			for _, memberID := range org.MemberIDs {
				if memberID == userID {
					return true, nil
				}
			}
			return false, nil
		}
		return false, handleCosmosError(err)
	}

	return true, nil
}

// AddMember adds a user as a member of an organization
func (r *OrganizationMemberRepository) AddMember(ctx context.Context, orgID, userID string) error {
	member := &types.OrganizationMember{
		ID:             orgID + "_" + userID,
		PartitionKey:   orgID,
		EntityType:     "organization_member",
		OrganizationID: orgID,
		UserID:         userID,
		CreatedAt:      time.Now(),
	}

	item, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal organization member: %w", err)
	}

	_, err = r.container.CreateItem(ctx, azcosmos.NewPartitionKeyString(member.PartitionKey), item, nil)
	return handleCosmosError(err)
}

// CreateOrganizationWithMember creates an organization and adds the creator as a member atomically.
// Uses a transactional batch since both are in the same container and partition.
// This ensures data consistency - either both succeed or both fail.
func (r *OrganizationMemberRepository) CreateOrganizationWithMember(ctx context.Context, org *types.Organization, userID string) error {
	// Set entity_type and ensure organization_id matches id
	org.EntityType = "organization"
	if org.OrganizationID == "" {
		org.OrganizationID = org.ID
	}
	org.PartitionKey = org.GetPartitionKey()

	// Create member record
	member := &types.OrganizationMember{
		ID:             org.ID + "_" + userID,
		PartitionKey:   org.OrganizationID,
		EntityType:     "organization_member",
		OrganizationID: org.ID,
		UserID:         userID,
		CreatedAt:      time.Now(),
	}

	// Marshal both items
	orgItem, err := json.Marshal(org)
	if err != nil {
		return fmt.Errorf("failed to marshal organization: %w", err)
	}

	memberItem, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal organization member: %w", err)
	}

	// Create transactional batch - both items are in the same partition
	batch := r.container.NewTransactionalBatch(azcosmos.NewPartitionKeyString(org.OrganizationID))
	batch.CreateItem(orgItem, nil)
	batch.CreateItem(memberItem, nil)

	// Execute batch - both succeed or both fail
	resp, err := r.container.ExecuteTransactionalBatch(ctx, batch, nil)
	if err != nil {
		return handleCosmosError(err)
	}

	if !resp.Success {
		return fmt.Errorf("transactional batch failed")
	}

	return nil
}

// RemoveMember removes a user from an organization
func (r *OrganizationMemberRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	memberID := orgID + "_" + userID
	partitionKey := orgID

	_, err := r.container.DeleteItem(ctx, azcosmos.NewPartitionKeyString(partitionKey), memberID, nil)
	return handleCosmosError(err)
}

// ListMembers lists all members of an organization
func (r *OrganizationMemberRepository) ListMembers(ctx context.Context, orgID string) ([]*types.OrganizationMember, error) {
	// Query only organization_member records in this partition
	query := "SELECT * FROM c WHERE c.entity_type = 'organization_member'"
	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(orgID), nil)

	var members []*types.OrganizationMember
	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range queryResponse.Items {
			var member types.OrganizationMember
			if err := json.Unmarshal(item, &member); err != nil {
				return nil, fmt.Errorf("failed to unmarshal organization member: %w", err)
			}
			// Double-check entity type
			if member.EntityType == "organization_member" {
				members = append(members, &member)
			}
		}
	}

	return members, nil
}

