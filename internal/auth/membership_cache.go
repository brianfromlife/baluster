package auth

import (
	"time"
)

// MembershipCache provides thread-safe in-memory caching for organization membership checks
type MembershipCache struct {
	cache *Cache[bool]
	ttl   time.Duration
}

// NewMembershipCache creates a new membership cache with the specified TTL
func NewMembershipCache(ttl time.Duration) *MembershipCache {
	return &MembershipCache{
		cache: NewCache[bool](),
		ttl:   ttl,
	}
}

// Get retrieves the membership status from cache
func (c *MembershipCache) Get(orgID, userID string) (bool, bool) {
	key := c.key(orgID, userID)
	return c.cache.Get(key)
}

// Set stores the membership status in cache
func (c *MembershipCache) Set(orgID, userID string, isMember bool) {
	key := c.key(orgID, userID)
	c.cache.SetWithTTL(key, isMember, c.ttl)
}

// Invalidate removes a specific membership entry from cache
func (c *MembershipCache) Invalidate(orgID, userID string) {
	key := c.key(orgID, userID)
	c.cache.Delete(key)
}

// InvalidateOrganization removes all membership entries for an organization
func (c *MembershipCache) InvalidateOrganization(orgID string) {
	prefix := orgID + ":"
	c.cache.InvalidatePrefix(prefix)
}

// key generates a cache key from orgID and userID
func (c *MembershipCache) key(orgID, userID string) string {
	return orgID + ":" + userID
}
