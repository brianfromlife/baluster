package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/brianfromlife/baluster/internal/types"
)

type ServiceKeyRepository struct {
	client    *Client
	container *azcosmos.ContainerClient
}

func NewServiceKeyRepository(client *Client) (*ServiceKeyRepository, error) {
	container, err := client.GetContainer("service_keys")
	if err != nil {
		return nil, err
	}
	return &ServiceKeyRepository{
		client:    client,
		container: container,
	}, nil
}

// HashToken hashes a token value for storage
func HashToken(tokenValue string) string {
	hash := sha256.Sum256([]byte(tokenValue))
	return hex.EncodeToString(hash[:])
}

// Create creates a new service key with audit history
func (r *ServiceKeyRepository) Create(ctx context.Context, token *types.ServiceKey, userID, githubID, username string) error {
	token.TokenValue = HashToken(token.TokenValue)
	token.PartitionKey = token.GetPartitionKey()

	// Create audit history record
	auditHistory := &types.AuditHistory{
		ID:                GenerateID(),
		OrganizationID:    token.OrganizationID,
		EntityType:        "audit_history",
		EntityID:          token.ID,
		Action:            types.AuditActionCreated,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
	}
	auditHistory.PartitionKey = auditHistory.GetPartitionKey()

	batch := r.container.NewTransactionalBatch(azcosmos.NewPartitionKeyString(token.PartitionKey))

	tokenItem, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	batch.CreateItem(tokenItem, nil)

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

// Get retrieves a service key by ID using the organization ID as the partition key
func (r *ServiceKeyRepository) Get(ctx context.Context, organizationID, id string) (*types.ServiceKey, error) {
	itemResponse, err := r.container.ReadItem(ctx, azcosmos.NewPartitionKeyString(organizationID), id, nil)
	if err != nil {
		return nil, handleCosmosError(err)
	}

	var token types.ServiceKey
	if err := json.Unmarshal(itemResponse.Value, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service key: %w", err)
	}

	// Verify entity type to ensure we got a service key, not audit history
	if token.EntityType != "" && token.EntityType != "service_key" {
		return nil, fmt.Errorf("token not found")
	}

	return &token, nil
}

// FindByTokenValue finds a service key by its hashed value (queries across all partitions)
func (r *ServiceKeyRepository) FindByTokenValue(ctx context.Context, tokenValue string) (*types.ServiceKey, error) {
	logger := r.client.logger.With(
		"function", "FindByTokenValue",
		"token_length", len(tokenValue),
	)
	logger.Info("searching for service key by token value")

	hashed := HashToken(tokenValue)
	logger = logger.With("hashed_token", hashed)
	logger.Debug("token hashed")

	query := fmt.Sprintf("SELECT * FROM c WHERE c.token_value = '%s' AND (c.entity_type = 'service_key' OR NOT IS_DEFINED(c.entity_type))", hashed)
	logger.Debug("executing query", "query", query)

	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKey(), nil)

	pageCount := 0
	for queryPager.More() {
		pageCount++
		logger.Debug("fetching next page", "page", pageCount)

		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			logger.Error("failed to fetch query page", "error", err, "page", pageCount)
			return nil, handleCosmosError(err)
		}

		logger.Debug("query page fetched", "page", pageCount, "item_count", len(queryResponse.Items))

		for i, item := range queryResponse.Items {
			var token types.ServiceKey
			if err := json.Unmarshal(item, &token); err != nil {
				logger.Error("failed to unmarshal token", "error", err, "item_index", i)
				return nil, fmt.Errorf("failed to unmarshal token: %w", err)
			}

			logger.Info("service key found",
				"service_key_id", token.ID,
				"organization_id", token.OrganizationID,
				"entity_type", token.EntityType,
			)
			return &token, nil
		}
	}

	logger.Warn("service key not found", "total_pages", pageCount)
	return nil, fmt.Errorf("token not found")
}

// FindByTokenValueInOrg finds a service key by its hashed value within a specific organization (more efficient)
func (r *ServiceKeyRepository) FindByTokenValueInOrg(ctx context.Context, organizationID, tokenValue string) (*types.ServiceKey, error) {
	logger := r.client.logger.With(
		"function", "FindByTokenValueInOrg",
		"organization_id", organizationID,
		"token_length", len(tokenValue),
	)
	logger.Info("searching for service key by token value in organization")

	hashed := HashToken(tokenValue)
	logger = logger.With("hashed_token", hashed)
	logger.Debug("token hashed")

	query := fmt.Sprintf("SELECT * FROM c WHERE c.organization_id = '%s' AND c.token_value = '%s' AND (c.entity_type = 'service_key' OR NOT IS_DEFINED(c.entity_type))", organizationID, hashed)
	logger.Debug("executing query", "query", query)

	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(organizationID), nil)

	pageCount := 0
	for queryPager.More() {
		pageCount++
		logger.Debug("fetching next page", "page", pageCount)

		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			logger.Error("failed to fetch query page", "error", err, "page", pageCount)
			return nil, handleCosmosError(err)
		}

		logger.Debug("query page fetched", "page", pageCount, "item_count", len(queryResponse.Items))

		for i, item := range queryResponse.Items {
			var token types.ServiceKey
			if err := json.Unmarshal(item, &token); err != nil {
				logger.Error("failed to unmarshal token", "error", err, "item_index", i)
				return nil, fmt.Errorf("failed to unmarshal token: %w", err)
			}

			logger.Info("service key found",
				"service_key_id", token.ID,
				"organization_id", token.OrganizationID,
				"entity_type", token.EntityType,
			)
			return &token, nil
		}
	}

	logger.Warn("service key not found in organization", "total_pages", pageCount)
	return nil, fmt.Errorf("token not found")
}

// CountByOrganization counts service keys for an organization
func (r *ServiceKeyRepository) CountByOrganization(ctx context.Context, organizationID string) (int, error) {
	query := fmt.Sprintf("SELECT VALUE COUNT(1) FROM c WHERE c.organization_id = '%s' AND (c.entity_type = 'service_key' OR NOT IS_DEFINED(c.entity_type))", organizationID)
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

// ListByOrganization lists service keys for an organization
func (r *ServiceKeyRepository) ListByOrganization(ctx context.Context, organizationID string) ([]*types.ServiceKey, error) {
	query := fmt.Sprintf("SELECT * FROM c WHERE c.organization_id = '%s' AND (c.entity_type = 'service_key' OR NOT IS_DEFINED(c.entity_type))", organizationID)
	queryPager := r.container.NewQueryItemsPager(query, azcosmos.NewPartitionKeyString(organizationID), nil)

	var tokens []*types.ServiceKey
	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, handleCosmosError(err)
		}

		for _, item := range queryResponse.Items {
			var token types.ServiceKey
			if err := json.Unmarshal(item, &token); err != nil {
				return nil, fmt.Errorf("failed to unmarshal token: %w", err)
			}
			tokens = append(tokens, &token)
		}
	}

	return tokens, nil
}

// Update updates a service key with audit history
func (r *ServiceKeyRepository) Update(ctx context.Context, token *types.ServiceKey, userID, githubID, username string) error {
	token.PartitionKey = token.GetPartitionKey()

	// Create audit history record
	auditHistory := &types.AuditHistory{
		ID:                GenerateID(),
		OrganizationID:    token.OrganizationID,
		EntityType:        "audit_history",
		EntityID:          token.ID,
		Action:            types.AuditActionUpdated,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
	}
	auditHistory.PartitionKey = auditHistory.GetPartitionKey()

	// Create transactional batch
	batch := r.container.NewTransactionalBatch(azcosmos.NewPartitionKeyString(token.PartitionKey))

	// Add service key update
	tokenItem, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	batch.ReplaceItem(token.ID, tokenItem, nil)

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

// Delete deletes a service key with audit history
func (r *ServiceKeyRepository) Delete(ctx context.Context, token *types.ServiceKey, userID, githubID, username string) error {
	token.PartitionKey = token.GetPartitionKey()

	// Create audit history record
	auditHistory := &types.AuditHistory{
		ID:                GenerateID(),
		OrganizationID:    token.OrganizationID,
		EntityType:        "audit_history",
		EntityID:          token.ID,
		Action:            types.AuditActionDeleted,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
	}
	auditHistory.PartitionKey = auditHistory.GetPartitionKey()

	// Create transactional batch
	batch := r.container.NewTransactionalBatch(azcosmos.NewPartitionKeyString(token.PartitionKey))

	// Add service key delete
	batch.DeleteItem(token.ID, nil)

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

// GetHistory retrieves audit history for a service key
func (r *ServiceKeyRepository) GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error) {
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
