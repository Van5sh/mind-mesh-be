package app

import (
	"context"
	"fmt"
	"os"

	"example/hello/internal/auth"
	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================
// App
// ============================================================

// App owns the application's shared infrastructure
// and all application dependencies.
type App struct {
	DB           *pgxpool.Pool
	Repositories Repositories
	Guards       Guards
	Services     Services
	OAuth        OAuthProviders
}

// ============================================================
// Repositories
// ============================================================

type Repositories struct {
	Activity  *repository.ActivityRepository
	Chat      *repository.ChatRepository
	File      *repository.FileRepository
	Flowchart *repository.FlowchartRepository
	Project   *repository.ProjectRepository
	Report    *repository.ReportRepository
	User      *repository.UserRepository
	OAuth     *repository.OAuthRepository
	Session   *repository.SessionRepository
}

// ============================================================
// Guards
// ============================================================

type Guards struct {
	Activity  *guards.ActivityGuard
	Chat      *guards.ChatGuard
	File      *guards.FileGuard
	Flowchart *guards.FlowchartGuard
	Project   *guards.ProjectGuard
	Report    *guards.ReportGuard
	User      *guards.UserGuard
}

// ============================================================
// Services
// ============================================================

type Services struct {
	Activity  *services.ActivityService
	Chat      *services.ChatService
	File      *services.FileService
	Flowchart *services.FlowchartService
	Project   *services.ProjectService
	Report    *services.ReportService
	User      *services.UserService
	Session   *services.SessionService

	// Authentication / OAuth service.
	Auth *auth.Service
}

// ============================================================
// OAuth Providers
// ============================================================

type OAuthProviders struct {
	Google *auth.GoogleProvider
	GitHub *auth.GitHubProvider
}

// ============================================================
// Constructor
// ============================================================

func New(
	ctx context.Context,
	databaseURL string,
) (*App, error) {

	// --------------------------------------------------------
	// Validate configuration
	// --------------------------------------------------------

	if databaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	// --------------------------------------------------------
	// PostgreSQL
	// --------------------------------------------------------

	db, err := pgxpool.New(
		ctx,
		databaseURL,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"create PostgreSQL pool: %w",
			err,
		)
	}

	// Verify the database connection.
	if err := db.Ping(ctx); err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"ping PostgreSQL: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// SQLC queries
	// --------------------------------------------------------

	queries := database.New(db)

	// --------------------------------------------------------
	// Repositories
	// --------------------------------------------------------

	repositories := Repositories{
		Activity: repository.NewActivityRepository(
			queries,
		),

		Chat: repository.NewChatRepository(
			queries,
		),

		File: repository.NewFileRepository(
			queries,
		),

		Flowchart: repository.NewFlowchartRepository(
			queries,
		),

		Project: repository.NewProjectRepository(
			db,
			queries,
		),

		Report: repository.NewReportRepository(
			queries,
		),

		User: repository.NewUserRepository(
			db,
			queries,
		),

		OAuth: repository.NewOAuthRepository(
			queries,
		),

		Session: repository.NewSessionRepository(
			queries,
		),
	}

	// --------------------------------------------------------
	// Guards
	// --------------------------------------------------------

	appGuards := Guards{
		Activity: guards.NewActivityGuard(
			repositories.Activity,
		),

		Chat: guards.NewChatGuard(
			repositories.Chat,
		),

		File: guards.NewFileGuard(
			repositories.File,
		),

		Flowchart: guards.NewFlowchartGuard(
			repositories.Flowchart,
		),

		Project: guards.NewProjectGuard(
			repositories.Project,
		),

		Report: guards.NewReportGuard(
			repositories.Report,
		),

		User: guards.NewUserGuard(
			repositories.User,
		),
	}

	// --------------------------------------------------------
	// Session Service
	// --------------------------------------------------------

	sessionService := services.NewSessionService(
		repositories.Session,
	)

	// --------------------------------------------------------
	// OAuth Providers
	// --------------------------------------------------------

	googleProvider := auth.NewGoogleProvider(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URL"),
	)

	githubProvider := auth.NewGitHubProvider(
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_SECRET"),
		os.Getenv("GITHUB_REDIRECT_URL"),
	)

	// --------------------------------------------------------
	// Auth Service
	// --------------------------------------------------------

	authService := auth.NewService(
		repositories.User,
		repositories.OAuth,
		sessionService,
		googleProvider,
		githubProvider,
	)

	// --------------------------------------------------------
	// Application
	// --------------------------------------------------------

	return &App{
		DB:           db,
		Repositories: repositories,
		Guards:       appGuards,

		OAuth: OAuthProviders{
			Google: googleProvider,
			GitHub: githubProvider,
		},

		Services: Services{
			Activity: services.NewActivityService(
				repositories.Activity,
				appGuards.Activity,
			),

			Chat: services.NewChatService(
				repositories.Chat,
				appGuards.Chat,
			),

			File: services.NewFileService(
				repositories.File,
				appGuards.File,
			),

			Flowchart: services.NewFlowchartService(
				repositories.Flowchart,
				appGuards.Flowchart,
			),

			Project: services.NewProjectService(
				repositories.Project,
				appGuards.User,
				appGuards.Project,
			),

			Report: services.NewReportService(
				repositories.Report,
				appGuards.Report,
			),

			User: services.NewUserService(
				repositories.User,
				appGuards.User,
			),

			Session: sessionService,

			Auth: authService,
		},
	}, nil
}

// ============================================================
// Close
// ============================================================

func (a *App) Close() {
	if a == nil {
		return
	}

	if a.DB != nil {
		a.DB.Close()
	}
}
