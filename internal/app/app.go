package app

import (
	"context"
	"fmt"
	"os"

	"example/hello/internal/database"
	"example/hello/internal/dbtrace"
	"example/hello/internal/guards"
	"example/hello/internal/realtime"
	"example/hello/internal/repository"
	"example/hello/internal/services"
	"example/hello/internal/services/ai"
	"example/hello/internal/services/aws"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB           *pgxpool.Pool
	Repositories Repositories
	Guards       Guards
	Services     Services
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
}

// ============================================================
// Constructor
// ============================================================

// Note: authentication (Firebase-backed OAuth + sessions) is constructed
// separately in cmd/server/main.go via auth.NewOAuthHandlerFromEnvironment,
// not here - it isn't part of App/Services because nothing in this package
// tree needs to read it back out; only main.go wires it into HTTP routes.

func New(
	ctx context.Context,
	databaseURL string,
	awsConfig *aws.AWSConfig,
) (*App, error) {

	// --------------------------------------------------------
	// Validate configuration
	// --------------------------------------------------------

	if databaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	if awsConfig == nil {
		return nil, fmt.Errorf(
			"AWS configuration is required",
		)
	}

	// --------------------------------------------------------
	// PostgreSQL
	// --------------------------------------------------------

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"parse DATABASE_URL: %w",
			err,
		)
	}

	// Dev aid: SQL_LOG=true logs every query's name (see internal/dbtrace).
	if os.Getenv("SQL_LOG") == "true" {
		poolConfig.ConnConfig.Tracer = dbtrace.Tracer{}
	}

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)

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
	// AWS-backed Services
	// --------------------------------------------------------

	s3Service := aws.NewS3Service(
		awsConfig.S3Client,
		awsConfig.S3Bucket,
	)

	sqsService := aws.NewSQSService(
		awsConfig.SQSClient,
		awsConfig.SQSQueueURL,
	)

	// --------------------------------------------------------
	// AI Service Client
	// --------------------------------------------------------

	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://localhost:8000"
	}

	aiClient := ai.NewClient(aiServiceURL)

	// --------------------------------------------------------
	// Real-time (in-process pub/sub, see internal/realtime)
	// --------------------------------------------------------

	chatBroker := realtime.NewChatBroker()

	// --------------------------------------------------------
	// Session Service
	// --------------------------------------------------------

	sessionService := services.NewSessionService(
		repositories.Session,
	)

	// --------------------------------------------------------
	// Application
	// --------------------------------------------------------

	return &App{
		DB: db,

		Repositories: repositories,

		Guards: appGuards,

		Services: Services{
			Activity: services.NewActivityService(
				repositories.Activity,
				appGuards.Activity,
			),

			Chat: services.NewChatService(
				repositories.Chat,
				appGuards.Chat,
				aiClient,
				chatBroker,
			),

			File: services.NewFileService(
				repositories.File,
				appGuards.File,
				s3Service,
				sqsService,
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
		},
	}, nil
}

func (a *App) Close() {
	if a == nil {
		return
	}

	if a.DB != nil {
		a.DB.Close()
	}
}
