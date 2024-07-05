package model

import (
	"time"

	"github.com/google/uuid"
)

type Part struct {
	ID        uuid.UUID
	ServerURL string
	UploadID  uuid.UUID
	Number    int32
	Size      int64
	CreatedAt time.Time
	Status    UploadStatus
}
