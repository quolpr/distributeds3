package model

type UploadStatus string

const (
	UploadStatusInProgress UploadStatus = "in_progress"
	UploadStatusDone       UploadStatus = "done"
)
