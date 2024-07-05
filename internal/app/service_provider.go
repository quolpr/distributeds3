package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quolpr/distributeds3/internal/config"
	"github.com/quolpr/distributeds3/internal/httpapi/upload"
	"github.com/quolpr/distributeds3/internal/queries/pg"
	"github.com/quolpr/distributeds3/internal/service/storage"
	"github.com/quolpr/distributeds3/internal/service/storage/repo/inmemstorage"
	uploadSvc "github.com/quolpr/distributeds3/internal/service/upload"
	"github.com/quolpr/distributeds3/internal/service/upload/repo"
	"github.com/quolpr/distributeds3/pkg/transaction"
)

type serviceProvider struct {
	Logger        *slog.Logger
	UploadHandler *upload.Handlers
}

func newServiceProvider(ctx context.Context) *serviceProvider {
	config, err := config.FromEnv()

	if err != nil {
		panic(err)
	}

	postgresPool, err := providePostgresql(ctx, config.DbURL)

	if err != nil {
		panic(err)
	}
	queries := pg.NewTxQueries(pg.New(postgresPool))

	storageService := storage.NewService(inmemstorage.NewInmemRepo())
	partRepo := repo.NewPartRepo(queries)
	uploadRepo := repo.NewUploadRepo(queries, partRepo)
	uploadService := uploadSvc.NewService(
		partRepo, uploadRepo, storageService,
		transaction.New(postgresPool),
	)

	return &serviceProvider{
		Logger:        slog.Default(),
		UploadHandler: upload.NewHandlers(uploadService),
	}
}

func providePostgresql(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	databaseConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error while parse postgreSQL's config | %w", err)
	}

	databaseConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	databaseConfig.AfterConnect = registerPostgresTypes

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

func registerPostgresTypes(ctx context.Context, conn *pgx.Conn) error {
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
