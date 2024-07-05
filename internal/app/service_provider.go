package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/quolpr/distributeds3/internal/config"
	"github.com/quolpr/distributeds3/internal/httpapi/upload"
	"github.com/quolpr/distributeds3/internal/queries/pg"
	"github.com/quolpr/distributeds3/internal/service/storage"
	"github.com/quolpr/distributeds3/internal/service/storage/repo/inmemstorage"
	uploadSvc "github.com/quolpr/distributeds3/internal/service/upload"
	"github.com/quolpr/distributeds3/internal/service/upload/repo"
	"github.com/quolpr/distributeds3/pkg/transaction"
	"github.com/quolpr/distributeds3/postgresql"
)

type serviceProvider struct {
	Logger        *slog.Logger
	UploadHandler *upload.Handlers
	UploadSvc     *uploadSvc.Service
}

func NewServiceProvider(ctx context.Context) (*serviceProvider, error) {
	config, err := config.FromEnv()

	if err != nil {
		return nil, fmt.Errorf("error while parse env config: %w", err)
	}

	postgresPool, err := providePostgresql(ctx, config.DBURL)

	if err != nil {
		return nil, fmt.Errorf("error while connect to postgreSQL: %w", err)
	}

	queries := pg.NewTxQueries(pg.New(postgresPool))
	storageService := storage.NewService(inmemstorage.NewInmemRepo())
	partRepo := repo.NewPartRepo(queries)
	uploadRepo := repo.NewUploadRepo(queries, partRepo)
	uploadService := uploadSvc.NewService(
		partRepo, uploadRepo, storageService,
		transaction.New(postgresPool),
	)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed set goose dialect: %w", err)
	}

	conn := stdlib.OpenDBFromPool(postgresPool)
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("Failed to close connection", "err", err)
		}
	}()

	goose.SetBaseFS(postgresql.EmbedMigrations)

	// NOTE: Usual migrations are migrating in different container on deploy
	// But for easier docker-compose up I managed to put it here
	if err := goose.Up(conn, "migrations"); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	connPgx, err := postgresPool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	err = RegisterPostgresTypes(ctx, connPgx.Conn())

	if err != nil {
		return nil, fmt.Errorf("failed to register postgres types: %w", err)
	}

	return &serviceProvider{
		Logger:        slog.Default(),
		UploadHandler: upload.NewHandlers(uploadService),
		UploadSvc:     uploadService,
	}, nil
}

func providePostgresql(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	databaseConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error while parse postgreSQL's config | %w", err)
	}

	databaseConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	pool, err := pgxpool.NewWithConfig(ctx, databaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error while connect to postgreSQL | %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while Ping postgreSQL | %w", err)
	}

	return pool, nil
}

func RegisterPostgresTypes(ctx context.Context, conn *pgx.Conn) error {
	dataTypeNames := map[string]any{
		"upload_status": pg.UploadStatus(""),
	}

	for typeName, val := range dataTypeNames {
		dataType, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return fmt.Errorf("failed to load pg type: %w", err)
		}

		conn.TypeMap().RegisterType(dataType)
		conn.TypeMap().RegisterDefaultPgType(val, typeName)
	}

	return nil
}
