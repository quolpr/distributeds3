package app

import "net/http"

func newRoutes(serviceProvider *serviceProvider) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /uploads", serviceProvider.UploadHandler.HandleUpload)
	mux.HandleFunc("GET /uploads/{id}", serviceProvider.UploadHandler.GetUpload)

	return mux
}
