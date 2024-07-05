package inmemstorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

var (
	ErrServerNotFound      = errors.New("server not found")
	ErrPartNotFound        = errors.New("part not found")
	ErrPartAlreadyUploaded = errors.New("part already uploaded")
)

type InmemRepo struct {
	// serverURL -> partID -> part
	parts map[string]map[uuid.UUID][]byte
}

func NewInmemRepo() *InmemRepo {
	parts := make(map[string]map[uuid.UUID][]byte)

	parts["http://localhost:8080"] = make(map[uuid.UUID][]byte)
	parts["http://localhost:8081"] = make(map[uuid.UUID][]byte)
	parts["http://localhost:8082"] = make(map[uuid.UUID][]byte)
	parts["http://localhost:8083"] = make(map[uuid.UUID][]byte)
	parts["http://localhost:8084"] = make(map[uuid.UUID][]byte)
	parts["http://localhost:8085"] = make(map[uuid.UUID][]byte)
	parts["http://localhost:8086"] = make(map[uuid.UUID][]byte)

	return &InmemRepo{parts: parts}
}

func (r *InmemRepo) GetAvailableServers(ctx context.Context) ([]string, error) {
	return maps.Keys(r.parts), nil
}

func (r *InmemRepo) UploadPart(ctx context.Context, partID uuid.UUID, serverURL string, reader io.Reader) error {
	m, ok := r.parts[serverURL]
	if !ok {
		return ErrServerNotFound
	}

	_, ok = m[partID]
	if ok {
		return ErrPartAlreadyUploaded
	}

	b, err := io.ReadAll(reader)

	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read part: %w", err)
	}

	m[partID] = b

	return nil
}

func (r *InmemRepo) GetPart(ctx context.Context, partID uuid.UUID, serverURL string) (io.Reader, error) {
	m, ok := r.parts[serverURL]
	if !ok {
		return nil, ErrServerNotFound
	}

	p, ok := m[partID]
	if !ok {
		return nil, ErrPartNotFound
	}

	return bytes.NewReader(p), nil
}

func (r *InmemRepo) CleanPart(ctx context.Context, partID uuid.UUID, serverURL string) error {
	m, ok := r.parts[serverURL]
	if !ok {
		return ErrServerNotFound
	}

	delete(m, partID)

	return nil
}

func (r *InmemRepo) ReadPart(ctx context.Context, id uuid.UUID, serverURL string, writer io.Writer) error {
	m, ok := r.parts[serverURL]
	if !ok {
		return ErrServerNotFound
	}

	p, ok := m[id]
	if !ok {
		return ErrPartNotFound
	}

	_, err := writer.Write(p)

	if err != nil {
		return fmt.Errorf("failed to write part: %w", err)
	}

	return nil
}
