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
	h := handlers.NewHandlers(repos, cfg, db)

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Request ID middleware (must be first)
	r.Use(middleware.RequestIDMiddleware())

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

	// Error handler middleware (must be last)
	r.Use(middleware.ErrorHandlerMiddleware())

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
				positions.GET("/:id/associations", middleware.RequireRole("admin"), h.Position.GetAssociations)
				positions.PUT("/:id", middleware.RequireRole("admin"), h.Position.Update)
				positions.DELETE("/:id", middleware.RequireRole("admin"), h.Position.Delete)
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
				branches.POST("/revenue/import", middleware.RequireRole("admin", "area_manager", "district_manager"), h.Branch.ImportRevenue)
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
				rotation.GET("/eligible-staff/:branchId", middleware.RequireRole("area_manager", "admin"), h.Rotation.GetEligibleStaff)
				// Schedule management (on/off days)
				rotation.POST("/schedule", middleware.RequireRole("area_manager", "district_manager", "admin"), h.Rotation.SetSchedule)
				rotation.GET("/schedule", h.Rotation.GetSchedules)
				rotation.PATCH("/schedule/:id", middleware.RequireRole("area_manager", "district_manager", "admin"), h.Rotation.UpdateSchedule)
				rotation.DELETE("/schedule/:id", middleware.RequireRole("area_manager", "district_manager", "admin"), h.Rotation.DeleteSchedule)
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

			// Admin test data generation (admin only)
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin"))
			{
				admin.POST("/test-data/generate-schedules", h.TestData.GenerateSchedules)
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
				quotas.POST("/import", middleware.RequireRole("admin", "area_manager"), h.Quota.Import)
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

			// Allocation criteria management (Admin only) - 5 criteria groups system
			allocationCriteria := protected.Group("/allocation-criteria")
			allocationCriteria.Use(middleware.RequireRole("admin"))
			{
				allocationCriteria.GET("/priority-order", h.AllocationCriteria.GetCriteriaPriorityOrder)
				allocationCriteria.PUT("/priority-order", h.AllocationCriteria.UpdateCriteriaPriorityOrder)
				allocationCriteria.POST("/priority-order/reset", h.AllocationCriteria.ResetCriteriaPriorityOrder)
			}

			// Specific Preferences (one of the 5 filters)
			specificPreferences := protected.Group("/specific-preferences")
			specificPreferences.Use(middleware.RequireRole("admin", "area_manager"))
			{
				specificPreferences.GET("", h.SpecificPreference.List)
				specificPreferences.POST("", h.SpecificPreference.Create)
				specificPreferences.GET("/:id", h.SpecificPreference.GetByID)
				specificPreferences.PUT("/:id", h.SpecificPreference.Update)
				specificPreferences.DELETE("/:id", h.SpecificPreference.Delete)
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
				scenarios.PUT("/:id/specific-staff-requirements", h.StaffRequirementScenario.UpdateSpecificStaffRequirements)
				scenarios.POST("/calculate", h.StaffRequirementScenario.CalculateRequirements)
				scenarios.POST("/match", h.StaffRequirementScenario.GetMatchingScenarios)
			}

			// Clinic-wide preferences management (Admin only)
			clinicPreferences := protected.Group("/clinic-preferences")
			clinicPreferences.Use(middleware.RequireRole("admin"))
			{
				clinicPreferences.GET("", h.ClinicWidePreference.List)
				clinicPreferences.POST("", h.ClinicWidePreference.Create)
				// Match route must come before :id routes to avoid route conflict
				clinicPreferences.GET("/match/:criteriaType", h.ClinicWidePreference.GetByCriteriaAndValue)
				clinicPreferences.GET("/:id", h.ClinicWidePreference.GetByID)
				clinicPreferences.PUT("/:id", h.ClinicWidePreference.Update)
				clinicPreferences.DELETE("/:id", h.ClinicWidePreference.Delete)
				clinicPreferences.POST("/:id/positions", h.ClinicWidePreference.AddPositionRequirement)
				clinicPreferences.PUT("/:id/positions/:positionId", h.ClinicWidePreference.UpdatePositionRequirement)
				clinicPreferences.DELETE("/:id/positions/:positionId", h.ClinicWidePreference.DeletePositionRequirement)
			}

			// Branch types management (Admin only)
			branchTypes := protected.Group("/branch-types")
			branchTypes.Use(middleware.RequireRole("admin"))
			{
				branchTypes.GET("", h.BranchType.List)
				branchTypes.POST("", h.BranchType.Create)
				branchTypes.GET("/:id", h.BranchType.GetByID)
				branchTypes.PUT("/:id", h.BranchType.Update)
				branchTypes.DELETE("/:id", h.BranchType.Delete)
				branchTypes.GET("/:id/requirements", h.BranchTypeRequirement.GetByBranchTypeID)
				branchTypes.POST("/:id/requirements", h.BranchTypeRequirement.Create)
				branchTypes.PUT("/:id/requirements/bulk", h.BranchTypeRequirement.BulkUpsert)
				branchTypes.GET("/:id/constraints", h.BranchTypeConstraints.GetByBranchTypeID)
				branchTypes.PUT("/:id/constraints", h.BranchTypeConstraints.UpdateConstraints)
			}

			// Staff groups management (Admin only)
			staffGroups := protected.Group("/staff-groups")
			staffGroups.Use(middleware.RequireRole("admin"))
			{
				staffGroups.GET("", h.StaffGroup.List)
				staffGroups.POST("", h.StaffGroup.Create)
				staffGroups.GET("/:id", h.StaffGroup.GetByID)
				staffGroups.PUT("/:id", h.StaffGroup.Update)
				staffGroups.DELETE("/:id", h.StaffGroup.Delete)
				staffGroups.POST("/:id/positions", h.StaffGroup.AddPosition)
				staffGroups.DELETE("/:id/positions/:positionId", h.StaffGroup.RemovePosition)
			}

			// Rotation staff branch position mapping (Admin only)
			rotationStaffMappings := protected.Group("/rotation-staff-branch-positions")
			rotationStaffMappings.Use(middleware.RequireRole("admin"))
			{
				rotationStaffMappings.GET("", h.RotationStaffBranchPosition.List)
				rotationStaffMappings.POST("", h.RotationStaffBranchPosition.Create)
				rotationStaffMappings.GET("/:id", h.RotationStaffBranchPosition.GetByID)
				rotationStaffMappings.PUT("/:id", h.RotationStaffBranchPosition.Update)
				rotationStaffMappings.DELETE("/:id", h.RotationStaffBranchPosition.Delete)
			}

			// Branch type requirements management (Admin only)
			branchTypeRequirements := protected.Group("/branch-type-requirements")
			branchTypeRequirements.Use(middleware.RequireRole("admin"))
			{
				branchTypeRequirements.PUT("/:id", h.BranchTypeRequirement.Update)
				branchTypeRequirements.DELETE("/:id", h.BranchTypeRequirement.Delete)
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
