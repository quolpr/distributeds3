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

type PartRepo struct {
	querier pg.Querier
	// qtx - querier для запуска в транзакционном режиме.
	qtx pg.QuerierTX
}

func NewPartRepo(querierTx pg.QuerierTX) *PartRepo {
	return &PartRepo{
		querier: querierTx,
		qtx:     querierTx,
	}
}

func (r *PartRepo) Create(ctx context.Context, part model.Part) error {
	err := r.querier.InsertPart(
		ctx,
		pg.InsertPartParams{
			ID:        part.ID,
			ServerUrl: part.ServerURL,
			UploadID:  part.UploadID,
			Number:    part.Number,
			Size:      part.Size,
			Status:    pg.UploadStatus(part.Status),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create part: %w", err)
	}

	return nil
}

func (r *PartRepo) GetParts(ctx context.Context, uploadID uuid.UUID) ([]model.Part, error) {
	rows, err := r.querier.GetUploadParts(
		ctx,
		uploadID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get parts: %w", err)
	}

	parts := make([]model.Part, len(rows))
	for i, r := range rows {
		parts[i] = model.Part{
			ID:        r.ID,
			ServerURL: r.ServerUrl,
			UploadID:  r.UploadID,
			Number:    r.Number,
			Size:      r.Size,
			CreatedAt: r.CreatedAt.Time,
			Status:    model.UploadStatus(r.Status),
		}
	}

	return parts, nil
}

func (r *PartRepo) MarkPartAsDone(ctx context.Context, partID uuid.UUID) error {
	err := r.querier.UpdatePartAsDone(
		ctx,
		partID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark part as done: %w", err)
	}

	return nil
}

func (r *PartRepo) GetOldInProgressParts(ctx context.Context, from time.Time) ([]model.Part, error) {
	time := pgtype.Timestamptz{
		Time:             from,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}

	rows, err := r.querier.GetOldInProgressParts(
		ctx,
		time,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get old in progress parts: %w", err)
	}

	parts := make([]model.Part, len(rows))
	for i, r := range rows {
		parts[i] = model.Part{
			ID:        r.ID,
			ServerURL: r.ServerUrl,
			UploadID:  r.UploadID,
			Number:    r.Number,
			Size:      r.Size,
			CreatedAt: r.CreatedAt.Time,
			Status:    model.UploadStatus(r.Status),
		}
	}

	return parts, nil
}

func (r *PartRepo) WithTx(tx pgx.Tx) *PartRepo {
	// если уже в транзакционном режиме - ничего не делаем
	if r.qtx == nil {
		return r
	}

	return &PartRepo{
		querier: r.qtx.WithTx(tx),
		qtx:     nil, // нельзя запускать транзакцию повторно
	}
}
