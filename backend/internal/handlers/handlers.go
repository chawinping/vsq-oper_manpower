package handlers

import (
	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type Handlers struct {
	Auth      *AuthHandler
	User      *UserHandler
	Staff     *StaffHandler
	Position  *PositionHandler
	Branch    *BranchHandler
	Schedule  *ScheduleHandler
	Rotation  *RotationHandler
	Settings  *SettingsHandler
	Dashboard *DashboardHandler
}

func NewHandlers(repos *postgres.Repositories, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:      NewAuthHandler(repos, cfg),
		User:      NewUserHandler(repos),
		Staff:     NewStaffHandler(repos),
		Position:  NewPositionHandler(repos),
		Branch:    NewBranchHandler(repos),
		Schedule:  NewScheduleHandler(repos),
		Rotation:  NewRotationHandler(repos, cfg),
		Settings:  NewSettingsHandler(repos),
		Dashboard: NewDashboardHandler(repos),
	}
}

