package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

var allowedThumbnailMimeTypes = [...]string{
	"image/jpeg",
	"image/png",
}

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse multipart form data", err)
		return
	}

	mpFile, header, err := r.FormFile("thumbnail")
	defer mpFile.Close()

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read file", err)
		return
	}

	contentType := header.Header.Get("Content-type")
	if !isMimeTypeAllowed(contentType) {
		respondWithError(w, http.StatusBadRequest, "Invalid mime type", err)
		return
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't read file", err)
		return
	}

	splitted := strings.Split(contentType, "/")

	ext := splitted[1]
	fileName := fmt.Sprintf("%v.%v", videoID, ext)

	fullPath := path.Join(cfg.assetsRoot, fileName)
	file, err := os.Create(fullPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create file", err)
		return
	}

	written, err := io.Copy(file, mpFile)
	defer file.Close()
	log.Println("written bytes:", written, "to", fullPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save file", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)

	if video.UserID != userID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%v/assets/%v", cfg.port, fileName)
	video = database.Video{
		ID:           video.ID,
		CreatedAt:    video.CreatedAt,
		UpdatedAt:    video.UpdatedAt,
		ThumbnailURL: &thumbnailURL,
		VideoURL:     video.VideoURL,
		CreateVideoParams: database.CreateVideoParams{
			Title:       video.Title,
			Description: video.Description,
			UserID:      video.UserID,
		},
	}

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}

func isMimeTypeAllowed(mime string) bool {
	for _, allowed := range allowedThumbnailMimeTypes {
		if allowed == mime {
			return true
		}
	}
	return false
}
