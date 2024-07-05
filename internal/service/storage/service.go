package storage

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/quolpr/distributeds3/internal/service/storage/repo"
)

type Service struct {
	repo repo.Repo
}

func NewService(repo repo.Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAvailableServers(ctx context.Context) ([]string, error) {
	return s.repo.GetAvailableServers(ctx) //nolint:wrapcheck
}

func (s *Service) CleanPart(ctx context.Context, id uuid.UUID, serverURL string) error {
	return s.repo.CleanPart(ctx, id, serverURL) //nolint:wrapcheck
}

func (s *Service) UploadPart(ctx context.Context, id uuid.UUID, serverURL string, reader io.Reader) error {
	return s.repo.UploadPart(ctx, id, serverURL, reader) //nolint:wrapcheck
}

func (s *Service) ReadPart(ctx context.Context, id uuid.UUID, serverURL string, writer io.Writer) error {
	return s.repo.ReadPart(ctx, id, serverURL, writer) //nolint:wrapcheck
}
