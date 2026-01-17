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
				// Branch configuration endpoints (more specific routes first)
				branches.GET("/:id/config/constraints", middleware.RequireRole("admin", "area_manager", "district_manager"), h.BranchConfig.GetConstraints)
				branches.PUT("/:id/config/constraints", middleware.RequireRole("admin", "area_manager", "district_manager"), h.BranchConfig.UpdateConstraints)
				branches.GET("/:id/config/weekly-revenue", middleware.RequireRole("admin", "area_manager", "district_manager"), h.BranchConfig.GetWeeklyRevenue)
				branches.PUT("/:id/config/weekly-revenue", middleware.RequireRole("admin", "area_manager", "district_manager"), h.BranchConfig.UpdateWeeklyRevenue)
				branches.GET("/:id/config/quotas", middleware.RequireRole("admin", "area_manager", "district_manager"), h.BranchConfig.GetQuotas)
				branches.PUT("/:id/config/quotas", middleware.RequireRole("admin", "area_manager", "district_manager"), h.BranchConfig.UpdateQuotas)
				branches.GET("/:id/config", middleware.RequireRole("admin", "area_manager", "district_manager"), h.BranchConfig.GetBranchConfig)
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
			effectiveBranches.PUT("/:id", h.EffectiveBranch.Update)
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
				// Zone and Branch management for Areas of Operation
				areasOfOperation.POST("/:id/zones", middleware.RequireRole("admin"), h.AreaOfOperation.AddZone)
				areasOfOperation.DELETE("/:id/zones/:zoneId", middleware.RequireRole("admin"), h.AreaOfOperation.RemoveZone)
				areasOfOperation.GET("/:id/zones", h.AreaOfOperation.GetZones)
				areasOfOperation.POST("/:id/branches", middleware.RequireRole("admin"), h.AreaOfOperation.AddBranch)
				areasOfOperation.DELETE("/:id/branches/:branchId", middleware.RequireRole("admin"), h.AreaOfOperation.RemoveBranch)
				areasOfOperation.GET("/:id/branches", h.AreaOfOperation.GetBranches)
				areasOfOperation.GET("/:id/all-branches", h.AreaOfOperation.GetAllBranches)
			}

			// Zone management (Master Data)
			zones := protected.Group("/zones")
			zones.Use(middleware.RequireRole("admin", "area_manager", "district_manager"))
			{
				zones.GET("", h.Zone.List)
				zones.GET("/:id", h.Zone.GetByID)
				zones.POST("", middleware.RequireRole("admin"), h.Zone.Create)
				zones.PUT("/:id", middleware.RequireRole("admin"), h.Zone.Update)
				zones.DELETE("/:id", middleware.RequireRole("admin"), h.Zone.Delete)
				zones.GET("/:id/branches", h.Zone.GetBranches)
				zones.PUT("/:id/branches", middleware.RequireRole("admin"), h.Zone.UpdateBranches)
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

			// Doctor management
			doctors := protected.Group("/doctors")
			{
				// Doctor CRUD
				doctors.GET("", middleware.RequireRole("admin", "area_manager"), h.Doctor.List)
				doctors.POST("", middleware.RequireRole("admin", "area_manager"), h.Doctor.Create)
				doctors.POST("/import", middleware.RequireRole("admin", "area_manager"), h.Doctor.Import)
				
				// Doctor Schedule (must be before /:id routes to avoid route conflict)
				doctors.GET("/:id/schedule", middleware.RequireRole("admin", "area_manager"), h.Doctor.GetMonthlySchedule)
				
				// Doctor CRUD by ID
				doctors.GET("/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.GetByID)
				doctors.PUT("/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.Update)
				doctors.DELETE("/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.Delete)
				doctors.POST("/assignments", middleware.RequireRole("admin", "area_manager", "branch_manager"), h.Doctor.CreateAssignment)
				doctors.GET("/assignments", h.Doctor.GetAssignments)
				doctors.DELETE("/assignments/:id", middleware.RequireRole("admin", "area_manager", "branch_manager"), h.Doctor.DeleteAssignment)
				
				// Doctor Preferences/Rules
				doctors.GET("/preferences", middleware.RequireRole("admin", "area_manager"), h.Doctor.ListPreferences)
				doctors.POST("/preferences", middleware.RequireRole("admin", "area_manager"), h.Doctor.CreatePreference)
				doctors.PUT("/preferences/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.UpdatePreference)
				doctors.DELETE("/preferences/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.DeletePreference)
				
				// Doctor On/Off Days
				doctors.POST("/on-off-days", middleware.RequireRole("admin", "branch_manager"), h.Doctor.CreateDoctorOnOffDay)
				doctors.GET("/on-off-days", h.Doctor.GetDoctorOnOffDays)
				doctors.DELETE("/on-off-days/:id", middleware.RequireRole("admin", "branch_manager"), h.Doctor.DeleteDoctorOnOffDay)
				
				// Doctor Default Schedules
				doctors.POST("/default-schedules", middleware.RequireRole("admin", "area_manager"), h.Doctor.CreateDefaultSchedule)
				doctors.POST("/default-schedules/import", middleware.RequireRole("admin", "area_manager"), h.Doctor.ImportDefaultSchedules)
				doctors.GET("/default-schedules", middleware.RequireRole("admin", "area_manager"), h.Doctor.GetDefaultSchedules)
				doctors.PUT("/default-schedules/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.UpdateDefaultSchedule)
				doctors.DELETE("/default-schedules/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.DeleteDefaultSchedule)
				
				// Doctor Weekly Off Days
				doctors.POST("/weekly-off-days", middleware.RequireRole("admin", "area_manager"), h.Doctor.CreateWeeklyOffDay)
				doctors.GET("/weekly-off-days", middleware.RequireRole("admin", "area_manager"), h.Doctor.GetWeeklyOffDays)
				doctors.DELETE("/weekly-off-days/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.DeleteWeeklyOffDay)
				
				// Doctor Schedule Overrides
				doctors.POST("/schedule-overrides", middleware.RequireRole("admin", "area_manager"), h.Doctor.CreateScheduleOverride)
				doctors.GET("/schedule-overrides", middleware.RequireRole("admin", "area_manager"), h.Doctor.GetScheduleOverrides)
				doctors.PUT("/schedule-overrides/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.UpdateScheduleOverride)
				doctors.DELETE("/schedule-overrides/:id", middleware.RequireRole("admin", "area_manager"), h.Doctor.DeleteScheduleOverride)
			}

			// Position quota management
			quotas := protected.Group("/quotas")
			{
				quotas.POST("", middleware.RequireRole("admin", "area_manager"), h.Quota.CreateQuota)
				quotas.GET("", h.Quota.GetQuotas)
				quotas.PUT("/:id", middleware.RequireRole("admin", "area_manager"), h.Quota.UpdateQuota)
				quotas.DELETE("/:id", middleware.RequireRole("admin", "area_manager"), h.Quota.DeleteQuota)
				quotas.GET("/branch/:branchId/status", h.Quota.GetBranchQuotaStatus)
			}

			// Overview endpoints
			overview := protected.Group("/overview")
			{
				overview.GET("/day", middleware.RequireRole("admin", "area_manager"), h.Overview.GetDayOverview)
				overview.GET("/monthly", h.Overview.GetMonthlyOverview)
			}

			// Allocation Report endpoints (TODO: Implement - Related: FR-RP-04)
			// reports := protected.Group("/reports")
			// {
			// 	reports.GET("", middleware.RequireRole("admin", "area_manager"), h.Report.GetReports)
			// 	reports.GET("/:id", middleware.RequireRole("admin", "area_manager"), h.Report.GetReport)
			// 	reports.POST("/generate", middleware.RequireRole("admin", "area_manager"), h.Report.GenerateReport)
			// 	reports.GET("/:id/export", middleware.RequireRole("admin", "area_manager"), h.Report.ExportReport)
			// }

			// Allocation criteria management (Admin only)
			allocationCriteria := protected.Group("/allocation-criteria")
			allocationCriteria.Use(middleware.RequireRole("admin"))
			{
				allocationCriteria.POST("", h.AllocationCriteria.CreateCriteria)
				allocationCriteria.GET("", h.AllocationCriteria.GetCriteria)
				allocationCriteria.GET("/:id", h.AllocationCriteria.GetCriteriaByID)
				allocationCriteria.PUT("/:id", h.AllocationCriteria.UpdateCriteria)
				allocationCriteria.DELETE("/:id", h.AllocationCriteria.DeleteCriteria)
			}

			// Revenue level tiers management (Admin only)
			revenueTiers := protected.Group("/revenue-level-tiers")
			revenueTiers.Use(middleware.RequireRole("admin"))
			{
				revenueTiers.GET("", h.RevenueLevelTier.List)
				revenueTiers.POST("", h.RevenueLevelTier.Create)
				revenueTiers.GET("/:id", h.RevenueLevelTier.GetByID)
				revenueTiers.PUT("/:id", h.RevenueLevelTier.Update)
				revenueTiers.DELETE("/:id", h.RevenueLevelTier.Delete)
				revenueTiers.POST("/match", h.RevenueLevelTier.GetTierForRevenue)
			}

			// Staff requirement scenarios management (Admin only)
			scenarios := protected.Group("/staff-requirement-scenarios")
			scenarios.Use(middleware.RequireRole("admin"))
			{
				scenarios.GET("", h.StaffRequirementScenario.List)
				scenarios.POST("", h.StaffRequirementScenario.Create)
				scenarios.GET("/:id", h.StaffRequirementScenario.GetByID)
				scenarios.PUT("/:id", h.StaffRequirementScenario.Update)
				scenarios.DELETE("/:id", h.StaffRequirementScenario.Delete)
				scenarios.PUT("/:id/position-requirements", h.StaffRequirementScenario.UpdatePositionRequirements)
				scenarios.POST("/calculate", h.StaffRequirementScenario.CalculateRequirements)
				scenarios.POST("/match", h.StaffRequirementScenario.GetMatchingScenarios)
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

