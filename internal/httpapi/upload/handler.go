package upload

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/quolpr/distributeds3/internal/service/upload"
)

const (
	maxUploadSize = 1024*1024*1024*10 + 1024
)

type Handlers struct {
	svc *upload.Service
}

func NewHandlers(svc *upload.Service) *Handlers {
	return &Handlers{
		svc: svc,
	}
}

type UploadResponse struct {
	UploadID string `json:"upload_id"`
}

type UploadErrorResponse struct {
	Error string `json:"error"`
}

func handleError(w http.ResponseWriter, err error) {
	slog.Error("Unable to handle request", "err", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	resp, err := json.Marshal(UploadErrorResponse{
		Error: err.Error(),
	})
	if err != nil {
		slog.Error("Unable to marshal error response", "err", err)
	}

	_, err = w.Write(resp)

	if err != nil {
		slog.Error("Unable to write response", "err", err)
	}
}

func (h *Handlers) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// function body of a http.HandlerFunc
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	reader, err := r.MultipartReader()

	if err != nil {
		handleError(w, err)

		return
	}

	fileSizeStr := make([]byte, 512) //nolint:gomnd
	p, err := reader.NextPart()

	if err != nil {
		handleError(w, err)

		return
	}

	if p.FormName() != "file_size" {
		handleError(w, errors.New("file_size is expected"))

		return
	}

	n, err := p.Read(fileSizeStr)
	if err != nil && !errors.Is(err, io.EOF) {
		handleError(w, err)

		return
	}

	fileSize, err := strconv.Atoi(string(fileSizeStr[:n]))
	if err != nil {
		handleError(w, err)

		return
	}

	p, err = reader.NextPart()
	if err != nil && !errors.Is(err, io.EOF) {
		handleError(w, err)

		return
	}

	if p.FormName() != "file" {
		handleError(w, errors.New("file is expected"))

		return
	}

	buf := bufio.NewReader(p)

	// TODO: parse file name
	// TODO: parse content type
	upload, err := h.svc.CreateUpload(r.Context(), int64(fileSize), "my-file", buf)

	if err != nil {
		handleError(w, err)

		return
	}

	jsonResponse, err := json.Marshal(UploadResponse{
		UploadID: upload.ID.String(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)

	if err != nil {
		slog.Error("Unable to write response", "err", err)
	}
}
func (h *Handlers) GetUpload(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)

	if err != nil {
		handleError(w, err)

		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	err = h.svc.ReadUpload(r.Context(), id, w)

	if err != nil {
		handleError(w, err)

		return
	}
}
