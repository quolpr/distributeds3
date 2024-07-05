package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/quolpr/distributeds3/internal/queries/pg"
	"github.com/quolpr/distributeds3/internal/service/upload/model"
)

type UploadRepo struct {
	querier pg.Querier
	// qtx - querier для запуска в транзакционном режиме.
	qtx pg.QuerierTX
}

func NewUploadRepo(querierTx pg.QuerierTX, partRepo *PartRepo) *UploadRepo {
	return &UploadRepo{
		querier: querierTx,
		qtx:     querierTx,
	}
}

func (r *UploadRepo) Create(ctx context.Context, upload model.Upload) error {
	err := r.querier.InsertUpload(
		ctx,
		pg.InsertUploadParams{
			ID:     upload.ID,
			Name:   upload.Name,
			Size:   upload.Size,
			Status: pg.UploadStatus(upload.Status),
			CreatedAt: pgtype.Timestamptz{
				Time:             upload.CreatedAt,
				InfinityModifier: pgtype.Finite,
				Valid:            true,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create upload: %w", err)
	}

	return nil
}

func (r *UploadRepo) MarkUploadAsDone(ctx context.Context, uploadID uuid.UUID) error {
	err := r.querier.UpdateUploadAsDone(
		ctx,
		uploadID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark upload as done: %w", err)
	}

	return nil
}
func (r *UploadRepo) GetOldInProgressUploads(ctx context.Context, from time.Time) ([]model.Upload, error) {
	time := pgtype.Timestamptz{
		Time:             from,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}

	rows, err := r.querier.GetOldInProgressUploads(
		ctx,
		time,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get old in progress uploads: %w", err)
	}

	uploads := make([]model.Upload, len(rows))
	for i, r := range rows {
		uploads[i] = model.Upload{
			ID:        r.ID,
			Name:      r.Name,
			Size:      r.Size,
			Status:    model.UploadStatus(r.Status),
			CreatedAt: r.CreatedAt.Time,
		}
	}

	return uploads, nil
}

func (r *UploadRepo) DeleteUploadByIDs(ctx context.Context, ids []uuid.UUID) error {
	err := r.querier.DeleteUploadsByIds(
		ctx,
		ids,
	)

	if err != nil {
		return fmt.Errorf("failed to delete uploads by ids: %w", err)
	}

	return nil
}

// func (r *UploadRepo) DeleteOldInProgressUploads(ctx context.Context, from time.Time) error {
// 	time := pgtype.Timestamptz{
// 		Time:             from,
// 		InfinityModifier: pgtype.Finite,
// 		Valid:            true,
// 	}
//
// 	return r.querier.DeleteOldInProgressUploads(
// 		ctx,
// 		time,
// 	)
// }

func (r *UploadRepo) WithTx(tx pgx.Tx) *UploadRepo {
	// если уже в транзакционном режиме - ничего не делаем
	if r.qtx == nil {
		return r
	}

	return &UploadRepo{
		querier: r.qtx.WithTx(tx),
		qtx:     nil, // нельзя запускать транзакцию повторно
	}
}
