package repo

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type Repo interface {
	GetAvailableServers(ctx context.Context) ([]string, error)
	UploadPart(ctx context.Context, id uuid.UUID, serverURL string, reader io.Reader) error
	ReadPart(ctx context.Context, id uuid.UUID, serverURL string, writer io.Writer) error
	CleanPart(ctx context.Context, id uuid.UUID, serverURL string) error
}
