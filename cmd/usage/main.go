package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type UploadResponse struct {
	UploadID string `json:"upload_id"`
}

func main() { //nolint:funlen
	filePath := "./test-file.txt"

	ctx := context.Background()
	url := "http://localhost:8080/uploads"
	method := "POST"

	fi, err := os.Stat(filePath)
	if err != nil {
		log.Println(err.Error())

		return
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("file_size", strconv.Itoa(int(fi.Size())))
	file, err := os.Open(filePath)

	if err != nil {
		log.Println(err.Error())

		return
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Println(err.Error())

		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Println(err.Error())

		return
	}

	defer func() {
		err := file.Close()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	part2, err := writer.CreateFormFile("file", filepath.Base("./test-file.txt"))

	if err != nil {
		log.Println(err.Error())

		return
	}

	_, err = io.Copy(part2, file)
	if err != nil {
		log.Println(err.Error())

		return
	}

	err = writer.Close()
	if err != nil {
		log.Println(err.Error())

		return
	}

	client := &http.Client{} //nolint:exhaustruct
	req, err := http.NewRequestWithContext(ctx, method, url, payload)

	if err != nil {
		log.Println(err.Error())

		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	log.Println("Uploading...")

	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())

		return
	}

	defer func() {
		err := res.Body.Close()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err.Error())

		return
	}

	log.Println("Got response from server: " + string(body))

	response := new(UploadResponse)
	err = json.Unmarshal(body, &response)

	if err != nil {
		log.Println(err.Error())

		return
	}

	log.Println("Getting file back...")

	req, err = http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf("http://localhost:8080/uploads/%s", response.UploadID), nil,
	)

	if err != nil {
		log.Println(err.Error())
	}

	res, err = client.Do(req)
	if err != nil {
		log.Println(err.Error())

		return
	}

	defer func() {
		err := res.Body.Close()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	body, err = io.ReadAll(res.Body)
	if err != nil {
		log.Println(err.Error())

		return
	}

	if string(fileContent) == string(body) {
		log.Println("Got file from server. Files are equal!")
	} else {
		log.Println("Got file from server. Files are not equal!")
	}
}
