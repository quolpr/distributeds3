package upload

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/quolpr/distributeds3/internal/service/storage"
	"github.com/quolpr/distributeds3/internal/service/upload/model"
	"github.com/quolpr/distributeds3/internal/service/upload/repo"
	"github.com/quolpr/distributeds3/pkg/transaction"
	"golang.org/x/exp/maps"
)

const (
	defaultParts         = 6
	defaultMaxUploadTime = time.Hour * 24
)

type Service struct {
	partRepo       *repo.PartRepo
	uploadRepo     *repo.UploadRepo
	storageService *storage.Service
	transaction    *transaction.Transaction

	parts         int64
	maxUploadTime time.Duration
}

func NewService(
	partRepo *repo.PartRepo, uploadRepo *repo.UploadRepo, storageService *storage.Service, tr *transaction.Transaction,
) *Service {
	return &Service{
		partRepo:       partRepo,
		uploadRepo:     uploadRepo,
		storageService: storageService,
		parts:          defaultParts,
		transaction:    tr,
		maxUploadTime:  defaultMaxUploadTime,
	}
}

func (s *Service) ReadUpload(ctx context.Context, id uuid.UUID, writer io.Writer) error {
	parts, err := s.partRepo.GetParts(ctx, id)

	if err != nil {
		return fmt.Errorf("unable to get parts: %w", err)
	}

	if len(parts) == 0 {
		return fmt.Errorf("no parts found")
	}

	for _, part := range parts {
		err := s.storageService.ReadPart(ctx, part.ID, part.ServerURL, writer)

		slog.Info("Reading part", "part", part, "parts", len(parts))

		if err != nil {
			return fmt.Errorf("unable to get part: %w", err)
		}
	}

	return nil
}

func (s *Service) CreateUpload(
	ctx context.Context, fileSize int64,
	fileName string, reader io.Reader,
) (model.Upload, error) {
	upload, parts, err := s.persistUpload(ctx, fileSize, fileName)

	if err != nil {
		return model.Upload{}, err
	}

	for _, part := range parts {
		reader := io.LimitReader(reader, part.Size)

		slog.Info("Uploading part", "part", part, "parts", len(parts))

		err := s.storageService.UploadPart(ctx, part.ID, part.ServerURL, reader)

		if err != nil {
			return model.Upload{}, fmt.Errorf("unable to upload part: %w", err)
		}

		err = s.partRepo.MarkPartAsDone(ctx, part.ID)

		if err != nil {
			return model.Upload{}, fmt.Errorf("unable to mark part as done: %w", err)
		}
	}

	err = s.uploadRepo.MarkUploadAsDone(ctx, upload.ID)

	if err != nil {
		return model.Upload{}, fmt.Errorf("unable to mark upload as done: %w", err)
	}

	return upload, nil
}

func (s *Service) CleanDangleUploads(ctx context.Context) error {
	parts, err := s.partRepo.GetOldInProgressParts(ctx, time.Now().Add(-s.maxUploadTime))

	if err != nil {
		return fmt.Errorf("unable to get old in progress uploads: %w", err)
	}

	for _, part := range parts {
		err := s.storageService.CleanPart(ctx, part.ID, part.ServerURL)

		if err != nil {
			return fmt.Errorf("unable to clean part: %w", err)
		}
	}

	ids := make(map[uuid.UUID]struct{})

	for _, part := range parts {
		ids[part.UploadID] = struct{}{}
	}

	// Due to cascade deletes in parts table, parts will be deleted too
	err = s.uploadRepo.DeleteUploadByIDs(ctx, maps.Keys(ids))

	if err != nil {
		return fmt.Errorf("unable to delete uploads: %w", err)
	}

	return nil
}

func (s *Service) persistUpload(
	ctx context.Context, fileSize int64, fileName string,
) (model.Upload, []model.Part, error) {
	upload := model.Upload{
		ID:        uuid.New(),
		Name:      fileName,
		Size:      fileSize,
		CreatedAt: time.Now(),
		Status:    model.UploadStatusInProgress,
	}
	parts := make([]model.Part, s.parts)

	servers, err := s.storageService.GetAvailableServers(ctx)

	if err != nil {
		return upload, parts, fmt.Errorf("unable to get available servers: %w", err)
	}

	randomServers, err := takeRandomServers(servers, int(s.parts))

	if err != nil {
		return upload, parts, err
	}

	for i := range s.parts {
		size := fileSize / s.parts

		if i == s.parts-1 {
			size += fileSize % s.parts
		}

		parts[i] = model.Part{
			ID:        uuid.New(),
			ServerURL: randomServers[i],
			UploadID:  upload.ID,
			Number:    int32(i),
			Size:      size,
			Status:    model.UploadStatusInProgress,
			CreatedAt: time.Now(),
		}
	}

	err = s.transaction.Exec(ctx, func(ctx context.Context, tx pgx.Tx) error {
		err := s.uploadRepo.WithTx(tx).Create(ctx, upload)
		if err != nil {
			return fmt.Errorf("unable to create upload: %w", err)
		}

		for i := range parts {
			err := s.partRepo.WithTx(tx).Create(ctx, parts[i])
			if err != nil {
				return fmt.Errorf("unable to create part: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return upload, parts, fmt.Errorf("unable to create upload: %w", err)
	}

	return upload, parts, nil
}

func takeRandomServers(initialServers []string, n int) ([]string, error) {
	if len(initialServers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	result := make([]string, 0, n)
	servers := make([]string, len(initialServers))
	copy(servers, initialServers)

	for range n {
		if len(servers) == 0 {
			servers = servers[:len(initialServers)]
		}

		idx := rand.Intn(len(servers)) //nolint:gosec
		server := servers[idx]

		servers[idx] = servers[len(servers)-1]
		servers = servers[:len(servers)-1]

		result = append(result, server)
	}

	return result, nil
}
