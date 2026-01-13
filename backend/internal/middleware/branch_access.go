package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireBranchAccess ensures branch managers can only access their own branch
func RequireBranchAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		
		// Only enforce for branch managers
		if role != "branch_manager" {
			c.Next()
			return
		}

		branchID := session.Get("branch_id")
		if branchID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Branch manager must be assigned to a branch"})
			c.Abort()
			return
		}

		branchIDStr, ok := branchID.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid branch ID in session"})
			c.Abort()
			return
		}

		// Parse and validate UUID
		userBranchID, err := uuid.Parse(branchIDStr)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid branch ID format"})
			c.Abort()
			return
		}

		// Set in context for handlers to use
		c.Set("user_branch_id", userBranchID)
		c.Next()
	}
}

// GetUserBranchID extracts the branch ID from context (set by RequireBranchAccess)
func GetUserBranchID(c *gin.Context) (*uuid.UUID, bool) {
	branchID, exists := c.Get("user_branch_id")
	if !exists {
		return nil, false
	}
	branchUUID, ok := branchID.(uuid.UUID)
	if !ok {
		return nil, false
	}
	return &branchUUID, true
}




