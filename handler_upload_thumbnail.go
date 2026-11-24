package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

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

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error on uploading video", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")

	vidMetaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "video not found", err)
		return
	}

	if vidMetaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "you are not authorized", errors.New("user dosent own the vid"))
		return
	}

	ext := "." + mediaType[6:]
	vidName := videoIDString + ext
	path := filepath.Join(cfg.assetsRoot, vidName)

	vidFile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "cannot create the file", err)
		return
	}

	_, err = io.Copy(vidFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating file", err)
		return
	}

	url := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, vidName)
	vidMetaData.ThumbnailURL = &url

	err = cfg.db.UpdateVideo(vidMetaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "cannot save the thumbnail", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vidMetaData)
}
