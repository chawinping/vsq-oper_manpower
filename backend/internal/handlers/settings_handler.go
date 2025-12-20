package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type SettingsHandler struct {
	repos *postgres.Repositories
}

func NewSettingsHandler(repos *postgres.Repositories) *SettingsHandler {
	return &SettingsHandler{repos: repos}
}

type UpdateSettingRequest struct {
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
}

func (h *SettingsHandler) GetAll(c *gin.Context) {
	settings, err := h.repos.Settings.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (h *SettingsHandler) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.repos.Settings.GetByKey(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if setting == nil {
		// Create new setting
		setting = &models.SystemSetting{
			Key:         key,
			Value:       req.Value,
			Description: req.Description,
		}
		if err := h.repos.Settings.Create(setting); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update existing setting
		setting.Value = req.Value
		if req.Description != "" {
			setting.Description = req.Description
		}
		if err := h.repos.Settings.Update(setting); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"setting": setting})
}



