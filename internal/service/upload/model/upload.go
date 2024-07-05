package model

import (
	"time"

	"github.com/google/uuid"
)

type Upload struct {
	ID        uuid.UUID
	Name      string
	Size      int64
	Status    UploadStatus
	CreatedAt time.Time
}
