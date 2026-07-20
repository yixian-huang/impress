package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/cache"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/pkg/apierror"
)

// rbacUserKey is the Gin context key for the fully loaded user (roles + perms).
// LoadRBACUser sets it once per request; RequirePermission only reads it.
const rbacUserKey = "rbac_user"

// loadUserWithCache loads a user with roles/permissions, using a short-lived cache
// to avoid hitting the DB on every admin request.
func loadUserWithCache(c *gin.Context, userID uint, userRepo repository.UserRepository, rbacCache *cache.Cache) (*model.User, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("user repository is not configured")
	}
	cacheKey := rbacCacheKey(userID)
	if rbacCache != nil {
		if cached, ok := rbacCache.Get(cacheKey); ok {
			user, valid := cached.(*model.User)
			if !valid {
				return nil, fmt.Errorf("invalid RBAC cache entry for user %d", userID)
			}
			return user, nil
		}
	}
	user, err := userRepo.FindByIDWithRoles(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}
	if rbacCache != nil {
		rbacCache.Set(cacheKey, user)
	}
	return user, nil
}

func rbacCacheKey(userID uint) string {
	return fmt.Sprintf("rbac:%d", userID)
}

// InvalidateRBACCache drops a user's permission cache entry (call after role changes).
func InvalidateRBACCache(rbacCache *cache.Cache, userID uint) {
	if rbacCache == nil || userID == 0 {
		return
	}
	rbacCache.Delete(rbacCacheKey(userID))
}

// LoadRBACUser loads the authenticated user with roles/permissions once and
// stores it on the request context. Mount once on /admin after Auth so
// per-route RequirePermission does not re-query the database.
func LoadRBACUser(userRepo repository.UserRepository, rbacCache *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx := GetUserContext(c)
		if userCtx == nil {
			respondWithError(c, apierror.Unauthorized("Authentication required"))
			return
		}
		user, err := loadUserWithCache(c, userCtx.UserID, userRepo, rbacCache)
		if err != nil {
			respondWithError(c, apierror.Forbidden("User not found"))
			return
		}
		c.Set(rbacUserKey, user)
		c.Next()
	}
}

// GetRBACUser returns the user loaded by LoadRBACUser (or set by RequirePermission fallback).
func GetRBACUser(c *gin.Context) *model.User {
	if c == nil {
		return nil
	}
	val, ok := c.Get(rbacUserKey)
	if !ok {
		return nil
	}
	user, ok := val.(*model.User)
	if !ok {
		return nil
	}
	return user
}

// resolveRBACUser returns the context user, or loads it once if missing
// (backward compatible with tests/modules that skip LoadRBACUser).
func resolveRBACUser(
	c *gin.Context,
	userRepo repository.UserRepository,
	rbacCache *cache.Cache,
) (*model.User, *apierror.APIError) {
	if user := GetRBACUser(c); user != nil {
		return user, nil
	}
	userCtx := GetUserContext(c)
	if userCtx == nil {
		return nil, apierror.Unauthorized("Authentication required")
	}
	user, err := loadUserWithCache(c, userCtx.UserID, userRepo, rbacCache)
	if err != nil {
		return nil, apierror.Forbidden("User not found")
	}
	c.Set(rbacUserKey, user)
	return user, nil
}

// RequirePermission checks resource:action against the RBAC user already on the
// context (preferred). Falls back to a single load if LoadRBACUser was not mounted.
func RequirePermission(resource, action string, userRepo repository.UserRepository, rbacCache *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, apiErr := resolveRBACUser(c, userRepo, rbacCache)
		if apiErr != nil {
			respondWithError(c, apiErr)
			return
		}
		if !user.HasRBACPermission(resource, action) {
			respondWithError(c, apierror.Forbidden(
				fmt.Sprintf("Permission denied: %s:%s", resource, action),
			))
			return
		}
		c.Next()
	}
}

// RequireAnyPermission returns middleware that checks if the user has at least
// one of the given permissions.
func RequireAnyPermission(perms []PermissionPair, userRepo repository.UserRepository, rbacCache *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, apiErr := resolveRBACUser(c, userRepo, rbacCache)
		if apiErr != nil {
			respondWithError(c, apiErr)
			return
		}
		for _, p := range perms {
			if user.HasRBACPermission(p.Resource, p.Action) {
				c.Next()
				return
			}
		}
		respondWithError(c, apierror.Forbidden("Permission denied: insufficient permissions"))
	}
}

// PermissionPair represents a resource:action permission pair
type PermissionPair struct {
	Resource string
	Action   string
}
