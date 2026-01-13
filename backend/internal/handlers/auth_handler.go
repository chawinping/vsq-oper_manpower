package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type AuthHandler struct {
	repos *postgres.Repositories
	cfg   *config.Config
}

func NewAuthHandler(repos *postgres.Repositories, cfg *config.Config) *AuthHandler {
	return &AuthHandler{repos: repos, cfg: cfg}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repos.User.GetByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	role, err := h.repos.Role.GetByID(user.RoleID)
	if err != nil || role == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID.String())
	session.Set("username", user.Username)
	session.Set("role", role.Name)
	if user.BranchID != nil {
		session.Set("branch_id", user.BranchID.String())
	}
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	response := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     role.Name,
	}
	if user.BranchID != nil {
		response["branch_id"] = user.BranchID
		// Fetch branch details to include branch name and code
		branch, err := h.repos.Branch.GetByID(*user.BranchID)
		if err != nil {
			// Log error but don't fail the request
			c.Error(err)
		} else if branch != nil {
			response["branch_name"] = branch.Name
			response["branch_code"] = branch.Code
		}
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("user_id")
	if userIDStr == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.repos.User.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	role, err := h.repos.Role.GetByID(user.RoleID)
	if err != nil || role == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	response := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     role.Name,
	}
	if user.BranchID != nil {
		response["branch_id"] = user.BranchID
		// Fetch branch details to include branch name and code
		branch, err := h.repos.Branch.GetByID(*user.BranchID)
		if err != nil {
			// Log error but don't fail the request
			c.Error(err)
		} else if branch != nil {
			response["branch_name"] = branch.Name
			response["branch_code"] = branch.Code
		}
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

func (h *AuthHandler) ListRoles(c *gin.Context) {
	roles, err := h.repos.Role.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// Helper function to hash password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

