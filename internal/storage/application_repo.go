package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApplicationRepository struct {
	client    *Client
	container *azcosmos.ContainerClient
}

func NewApplicationRepository(client *Client) (*ApplicationRepository, error) {
	container, err := client.GetContainer("applications")
	if err != nil {
		return nil, err
	}
	return &ApplicationRepository{
		client:    client,
		container: container,
	}, nil
}

// Create creates a new application with audit history
func (r *ApplicationRepository) Create(ctx context.Context, app *types.Application, userID, githubID, username string) error {
	app.PartitionKey = app.GetPartitionKey()

	auditHistory := &types.AuditHistory{
		ID:                GenerateID(),
		OrganizationID:    app.OrganizationID,
		EntityType:        "audit_history",
		EntityID:          app.ID,
		Action:            types.AuditActionCreated,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
	}

	auditHistory.PartitionKey = auditHistory.GetPartitionKey()

	batch := r.container.NewTransactionalBatch(azcosmos.NewPartitionKeyString(app.PartitionKey))

	appItem, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("failed to marshal application: %w", err)
	}

	batch.CreateItem(appItem, nil)

	auditItem, err := json.Marshal(auditHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal audit history: %w", err)
	}
	batch.CreateItem(auditItem, nil)

	resp, err := r.container.ExecuteTransactionalBatch(ctx, batch, nil)
	if err != nil {
		return handleCosmosError(err)
	}

	if !resp.Success {
		return handleCosmosError(fmt.Errorf("batch operation failed"))
	}

	return nil
}

// CountByOrganization counts applications for an organization
func (r *ApplicationRepository) CountByOrganization(ctx context.Context, organizationID string) (int, error) {
	query := fmt.Sprintf("SELECT VALUE COUNT(1) FROM c WHERE c.organization_id = '%s' AND (c.entity_type = 'application' OR NOT IS_DEFINED(c.entity_type))", organizationID)
	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(organizationID), nil)

	if !queryPager.More() {
		return 0, nil
	}

	queryResponse, err := queryPager.NextPage(ctx)
	if err != nil {
		return 0, handleCosmosError(err)
	}

	if len(queryResponse.Items) == 0 {
		return 0, nil
	}

	var result float64
	if err := json.Unmarshal(queryResponse.Items[0], &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal count: %w", err)
	}

	return int(result), nil
}

// ListByOrganization lists applications for an organization
func (r *ApplicationRepository) ListByOrganization(ctx context.Context, organizationID string) ([]*types.Application, error) {
	query := fmt.Sprintf("SELECT * FROM c WHERE c.organization_id = '%s' AND (c.entity_type = 'application' OR NOT IS_DEFINED(c.entity_type))", organizationID)
	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(organizationID), nil)

	var apps []*types.Application
	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range queryResponse.Items {
			var app types.Application
			if err := json.Unmarshal(item, &app); err != nil {
				return nil, fmt.Errorf("failed to unmarshal application: %w", err)
			}
			apps = append(apps, &app)
		}
	}

	return apps, nil
}

// Get retrieves an application by ID using the organization ID as the partition key
func (r *ApplicationRepository) Get(ctx context.Context, organizationID, id string) (*types.Application, error) {
	itemResponse, err := r.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(organizationID), id, nil)
	if err != nil {
		return nil, handleCosmosError(err)
	}

	var app types.Application
	if err := json.Unmarshal(itemResponse.Value, &app); err != nil {
		return nil, fmt.Errorf("failed to unmarshal application: %w", err)
	}

	return &app, nil
}

// Update updates an application with audit history
func (r *ApplicationRepository) Update(ctx context.Context, app *types.Application, userID, githubID, username string) error {
	app.PartitionKey = app.GetPartitionKey()

	// Create audit history record
	auditHistory := &types.AuditHistory{
		ID:                GenerateID(),
		OrganizationID:    app.OrganizationID,
		EntityType:        "audit_history",
		EntityID:          app.ID,
		Action:            types.AuditActionUpdated,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
	}
	auditHistory.PartitionKey = auditHistory.GetPartitionKey()

	// Create transactional batch
	batch := r.container.NewTransactionalBatch(azcosmos.NewPartitionKeyString(app.PartitionKey))

	// Add application update
	appItem, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("failed to marshal application: %w", err)
	}
	batch.ReplaceItem(app.ID, appItem, nil)

	// Add audit history create
	auditItem, err := json.Marshal(auditHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal audit history: %w", err)
	}
	batch.CreateItem(auditItem, nil)

	// Execute batch
	resp, err := r.container.ExecuteTransactionalBatch(ctx, batch, nil)
	if err != nil {
		return handleCosmosError(err)
	}

	if !resp.Success {
		return handleCosmosError(fmt.Errorf("batch operation failed"))
	}

	return nil
}

// Delete deletes an application with audit history
func (r *ApplicationRepository) Delete(ctx context.Context, app *types.Application, userID, githubID, username string) error {
	app.PartitionKey = app.GetPartitionKey()

	auditHistory := &types.AuditHistory{
		ID:                GenerateID(),
		OrganizationID:    app.OrganizationID,
		EntityType:        "audit_history",
		EntityID:          app.ID,
		Action:            types.AuditActionDeleted,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
	}
	auditHistory.PartitionKey = auditHistory.GetPartitionKey()

	// Create transactional batch
	batch := r.container.NewTransactionalBatch(azcosmos.NewPartitionKeyString(app.PartitionKey))

	// Add application delete
	batch.DeleteItem(app.ID, nil)

	// Add audit history create
	auditItem, err := json.Marshal(auditHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal audit history: %w", err)
	}
	batch.CreateItem(auditItem, nil)

	// Execute batch
	resp, err := r.container.ExecuteTransactionalBatch(ctx, batch, nil)
	if err != nil {
		return handleCosmosError(err)
	}

	if !resp.Success {
		return handleCosmosError(fmt.Errorf("batch operation failed"))
	}

	return nil
}

// GetHistory retrieves audit history for an application
func (r *ApplicationRepository) GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error) {
	query := fmt.Sprintf("SELECT * FROM c WHERE c.organization_id = '%s' AND c.entity_type = 'audit_history' AND c.entity_id = '%s' ORDER BY c.created_at DESC", organizationID, entityID)
	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(organizationID), nil)

	var history []*types.AuditHistory
	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range queryResponse.Items {
			var audit types.AuditHistory
			if err := json.Unmarshal(item, &audit); err != nil {
				return nil, fmt.Errorf("failed to unmarshal audit history: %w", err)
			}
			history = append(history, &audit)
		}
	}

	return history, nil
}
