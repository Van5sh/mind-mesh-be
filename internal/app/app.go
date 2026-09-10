package app

import (
	"context"
	"fmt"

	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

// App owns the application's shared infrastructure and dependencies.
type App struct {
	DB           *pgxpool.Pool
	Repositories Repositories
	Guards       Guards
	Services     Services
}

type Repositories struct {
	Activity  *repository.ActivityRepository
	Chat      *repository.ChatRepository
	File      *repository.FileRepository
	Flowchart *repository.FlowchartRepository
	Project   *repository.ProjectRepository
	Report    *repository.ReportRepository
	User      *repository.UserRepository
}

type Guards struct {
	Activity  *guards.ActivityGuard
	Chat      *guards.ChatGuard
	File      *guards.FileGuard
	Flowchart *guards.FlowchartGuard
	Project   *guards.ProjectGuard
	Report    *guards.ReportGuard
	User      *guards.UserGuard
}

type Services struct {
	Activity  *services.ActivityService
	Chat      *services.ChatService
	File      *services.FileService
	Flowchart *services.FlowchartService
	Project   *services.ProjectService
	Report    *services.ReportService
	User      *services.UserService
}

func New(ctx context.Context, databaseURL string) (*App, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	queries := database.New(db)
	repositories := Repositories{
		Activity:  repository.NewActivityRepository(queries),
		Chat:      repository.NewChatRepository(queries),
		File:      repository.NewFileRepository(queries),
		Flowchart: repository.NewFlowchartRepository(queries),
		Project:   repository.NewProjectRepository(db, queries),
		Report:    repository.NewReportRepository(queries),
		User:      repository.NewUserRepository(db, queries),
	}
	appGuards := Guards{
		Activity:  guards.NewActivityGuard(repositories.Activity),
		Chat:      guards.NewChatGuard(repositories.Chat),
		File:      guards.NewFileGuard(repositories.File),
		Flowchart: guards.NewFlowchartGuard(repositories.Flowchart),
		Project:   guards.NewProjectGuard(repositories.Project),
		Report:    guards.NewReportGuard(repositories.Report),
		User:      guards.NewUserGuard(repositories.User),
	}

	return &App{
		DB:           db,
		Repositories: repositories,
		Guards:       appGuards,
		Services: Services{
			Activity:  services.NewActivityService(repositories.Activity, appGuards.Activity),
			Chat:      services.NewChatService(repositories.Chat, appGuards.Chat),
			File:      services.NewFileService(repositories.File, appGuards.File),
			Flowchart: services.NewFlowchartService(repositories.Flowchart, appGuards.Flowchart),
			Project:   services.NewProjectService(repositories.Project, appGuards.User, appGuards.Project),
			Report:    services.NewReportService(repositories.Report, appGuards.Report),
			User:      services.NewUserService(repositories.User, appGuards.User),
		},
	}, nil
}

func (a *App) Close() {
	if a != nil && a.DB != nil {
		a.DB.Close()
	}
}
