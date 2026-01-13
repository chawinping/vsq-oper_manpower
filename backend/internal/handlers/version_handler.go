package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type VersionHandler struct{}

func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
	BuildTime string `json:"buildTime"`
}

type VersionResponse struct {
	Frontend VersionInfo `json:"frontend"`
	Backend  VersionInfo `json:"backend"`
	Database VersionInfo `json:"database"`
}

func (h *VersionHandler) GetVersion(c *gin.Context) {
	// Try multiple possible paths for version files
	// This handles different execution contexts (local dev, Docker, etc.)
	// In Docker, files are in the working directory (/root/)
	// In local dev, files are in backend/ directory
	
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	// Get executable directory (more reliable for finding files)
	execPath, err := os.Executable()
	var execDir string
	if err == nil {
		execDir = filepath.Dir(execPath)
	} else {
		execDir = wd
	}

	// Build list of possible paths, ordered by likelihood
	possibleBackendPaths := []string{
		filepath.Join(execDir, "VERSION.json"),      // Next to executable (Docker: /root/VERSION.json)
		filepath.Join(wd, "VERSION.json"),           // Current working directory
		"VERSION.json",                               // Current directory (relative)
		"./VERSION.json",                             // Explicit current directory
		"/root/VERSION.json",                         // Docker container default location
		filepath.Join(wd, "backend", "VERSION.json"), // From project root (if running from root)
		"backend/VERSION.json",                       // Relative from project root
	}

	possibleDBPaths := []string{
		filepath.Join(execDir, "DATABASE_VERSION.json"),      // Next to executable (Docker: /root/DATABASE_VERSION.json)
		filepath.Join(wd, "DATABASE_VERSION.json"),           // Current working directory
		"DATABASE_VERSION.json",                               // Current directory (relative)
		"./DATABASE_VERSION.json",                             // Explicit current directory
		"/root/DATABASE_VERSION.json",                         // Docker container default location
		filepath.Join(wd, "backend", "DATABASE_VERSION.json"), // From project root (if running from root)
		"backend/DATABASE_VERSION.json",                       // Relative from project root
	}

	var backendVersion, dbVersion VersionInfo
	var backendFound, dbFound bool

	// Find and read backend version
	for _, path := range possibleBackendPaths {
		if _, err := os.Stat(path); err == nil {
			var backendErr error
			backendVersion, backendErr = readVersionFile(path)
			if backendErr == nil {
				backendFound = true
				// Log in development mode only
				if os.Getenv("GIN_MODE") != "release" {
					log.Printf("Found backend version file at: %s", path)
				}
				break
			} else if os.Getenv("GIN_MODE") != "release" {
				log.Printf("Failed to read backend version file at %s: %v", path, backendErr)
			}
		}
	}

	// Find and read database version
	for _, path := range possibleDBPaths {
		if _, err := os.Stat(path); err == nil {
			var dbErr error
			dbVersion, dbErr = readVersionFile(path)
			if dbErr == nil {
				dbFound = true
				// Log in development mode only
				if os.Getenv("GIN_MODE") != "release" {
					log.Printf("Found database version file at: %s", path)
				}
				break
			} else if os.Getenv("GIN_MODE") != "release" {
				log.Printf("Failed to read database version file at %s: %v", path, dbErr)
			}
		}
	}

	// Log if files not found (development mode only)
	if os.Getenv("GIN_MODE") != "release" {
		if !backendFound {
			log.Printf("Backend version file not found. Tried paths: %v", possibleBackendPaths)
			log.Printf("Current working directory: %s, Executable directory: %s", wd, execDir)
		}
		if !dbFound {
			log.Printf("Database version file not found. Tried paths: %v", possibleDBPaths)
		}
	}

	// If still not found, set defaults
	if !backendFound {
		backendVersion = VersionInfo{
			Version:   "unknown",
			BuildDate: "unknown",
			BuildTime: "unknown",
		}
	}

	if !dbFound {
		dbVersion = VersionInfo{
			Version:   "unknown",
			BuildDate: "unknown",
			BuildTime: "unknown",
		}
	}

	// Frontend version is loaded by frontend itself from /VERSION.json
	// The frontend will fetch this separately
	frontendVersion := VersionInfo{
		Version:   "loaded-by-frontend",
		BuildDate: "loaded-by-frontend",
		BuildTime: "loaded-by-frontend",
	}

	response := VersionResponse{
		Frontend: frontendVersion,
		Backend:  backendVersion,
		Database: dbVersion,
	}

	c.JSON(http.StatusOK, response)
}

func readVersionFile(filePath string) (VersionInfo, error) {
	var version VersionInfo

	data, err := os.ReadFile(filePath)
	if err != nil {
		return version, err
	}

	// Remove UTF-8 BOM if present (0xEF, 0xBB, 0xBF)
	// BOM can cause "invalid character 'ï' looking for beginning of value" error
	dataStr := string(data)
	if strings.HasPrefix(dataStr, "\ufeff") {
		dataStr = strings.TrimPrefix(dataStr, "\ufeff")
		data = []byte(dataStr)
	} else if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		// Handle BOM as bytes (more reliable)
		data = data[3:]
	}

	if err := json.Unmarshal(data, &version); err != nil {
		return version, err
	}

	return version, nil
}

