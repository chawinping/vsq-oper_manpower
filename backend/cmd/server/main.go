package main

import (
	"log"
	"net/http"
	"os"

	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/handlers"
	"vsq-oper-manpower/backend/internal/middleware"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	repos := postgres.NewRepositories(db)

	// Initialize handlers
	h := handlers.NewHandlers(repos, cfg)

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(middleware.CORS(cfg))

	// Session middleware
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	// For development: Use SameSite Lax (works with localhost)
	// For production: Use SameSite None with Secure true (requires HTTPS)
	isDevelopment := os.Getenv("ENVIRONMENT") != "production"
	sameSite := http.SameSiteLaxMode
	secure := false
	if !isDevelopment {
		sameSite = http.SameSiteNoneMode
		secure = true
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		// Don't set Domain for localhost - browsers handle it automatically
	})
	r.Use(sessions.Sessions("vsq_session", store))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Version endpoint (public)
		v1 := api.Group("/v1")
		{
			v1.GET("/version", h.Version.GetVersion)
		}

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.Auth.Login)
			auth.POST("/logout", h.Auth.Logout)
			auth.GET("/me", middleware.RequireAuth(), h.Auth.Me)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.RequireAuth())
		{
			// User management (admin only)
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin"))
			{
				users.GET("", h.User.List)
				users.POST("", h.User.Create)
				users.PUT("/:id", h.User.Update)
				users.DELETE("/:id", h.User.Delete)
			}

			// Roles
			roles := protected.Group("/roles")
			{
				roles.GET("", h.Auth.ListRoles)
			}

			// Positions
			positions := protected.Group("/positions")
			{
				positions.GET("", h.Position.List)
				positions.GET("/:id", h.Position.GetByID)
				positions.PUT("/:id", middleware.RequireRole("admin"), h.Position.Update)
			}

			// Staff management
			staff := protected.Group("/staff")
			staff.Use(middleware.RequireBranchAccess())
			{
				staff.GET("", h.Staff.List)
				staff.POST("", middleware.RequireRole("admin", "area_manager", "district_manager", "branch_manager"), h.Staff.Create)
				staff.PUT("/:id", middleware.RequireRole("admin", "area_manager", "district_manager", "branch_manager"), h.Staff.Update)
				staff.DELETE("/:id", middleware.RequireRole("admin"), h.Staff.Delete)
				staff.POST("/import", middleware.RequireRole("admin", "area_manager", "district_manager"), h.Staff.Import)
			}

			// Branch management
			branches := protected.Group("/branches")
			{
				branches.GET("", h.Branch.List)
				branches.POST("", middleware.RequireRole("admin", "area_manager", "district_manager"), h.Branch.Create)
				branches.PUT("/:id", middleware.RequireRole("admin", "area_manager", "district_manager"), h.Branch.Update)
				branches.DELETE("/:id", middleware.RequireRole("admin"), h.Branch.Delete)
				branches.GET("/:id/revenue", h.Branch.GetRevenue)
			}

			// Staff scheduling
			schedules := protected.Group("/schedules")
			schedules.Use(middleware.RequireBranchAccess())
			{
				schedules.GET("/branch/:branchId", h.Schedule.GetBranchSchedule)
				schedules.POST("", middleware.RequireRole("branch_manager", "admin"), h.Schedule.Create)
				schedules.GET("/monthly", h.Schedule.GetMonthlyView)
			}

			// Rotation staff scheduling
			rotation := protected.Group("/rotation")
			{
				rotation.GET("/assignments", h.Rotation.GetAssignments)
				rotation.POST("/assign", middleware.RequireRole("area_manager", "district_manager", "admin"), h.Rotation.Assign)
				rotation.POST("/bulk-assign", middleware.RequireRole("area_manager", "admin"), h.Rotation.BulkAssign)
				rotation.DELETE("/assign/:id", middleware.RequireRole("area_manager", "district_manager", "admin"), h.Rotation.RemoveAssignment)
				rotation.GET("/suggestions", middleware.RequireRole("area_manager", "district_manager", "admin"), h.Rotation.GetSuggestions)
				rotation.POST("/regenerate-suggestions", middleware.RequireRole("area_manager", "district_manager", "admin"), h.Rotation.RegenerateSuggestions)
				rotation.GET("/eligible-staff/:branchId", middleware.RequireRole("area_manager", "admin"), h.Rotation.GetEligibleStaff)
			}

			// Effective branches management
			effectiveBranches := protected.Group("/effective-branches")
			effectiveBranches.Use(middleware.RequireRole("admin", "area_manager", "district_manager"))
			{
				effectiveBranches.GET("/rotation-staff/:rotationStaffId", h.EffectiveBranch.GetByRotationStaffID)
				effectiveBranches.POST("", h.EffectiveBranch.Create)
				effectiveBranches.DELETE("/:id", h.EffectiveBranch.Delete)
				effectiveBranches.PUT("/bulk-update", h.EffectiveBranch.BulkUpdate)
			}

			// Areas of Operation management (Master Data)
			areasOfOperation := protected.Group("/areas-of-operation")
			areasOfOperation.Use(middleware.RequireRole("admin", "area_manager", "district_manager"))
			{
				areasOfOperation.GET("", h.AreaOfOperation.List)
				areasOfOperation.GET("/:id", h.AreaOfOperation.GetByID)
				areasOfOperation.POST("", middleware.RequireRole("admin"), h.AreaOfOperation.Create)
				areasOfOperation.PUT("/:id", middleware.RequireRole("admin"), h.AreaOfOperation.Update)
				areasOfOperation.DELETE("/:id", middleware.RequireRole("admin"), h.AreaOfOperation.Delete)
			}

			// System settings
			settings := protected.Group("/settings")
			{
				settings.GET("", middleware.RequireRole("admin"), h.Settings.GetAll)
				settings.PUT("/:key", middleware.RequireRole("admin"), h.Settings.Update)
			}

			// Dashboard
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("", h.Dashboard.GetOverview)
			}
		}
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

