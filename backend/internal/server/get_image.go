package mediaserver

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/server/emby_jellyfin"
	"aura/internal/server/plex"
	mediaserver_shared "aura/internal/server/shared"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
)

func GetImageFromMediaServer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	ratingKey := chi.URLParam(r, "ratingKey")
	imageType := chi.URLParam(r, "imageType")
	if ratingKey == "" || imageType == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("missing rating key or image type"),
			Log: logging.Log{
				Message: "Missing rating key or image type in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	} else if imageType != "poster" && imageType != "backdrop" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: fmt.Errorf("invalid image type in URL"),
			Log: logging.Log{
				Message: "Invalid image type in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	var mediaServer mediaserver_shared.MediaServer
	var tmpFolder string
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
		tmpFolder = plex.PlexTempImageFolder
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
		tmpFolder = emby_jellyfin.EmbyJellyTempImageFolder
	default:
		logErr := logging.ErrorLog{
			Err: fmt.Errorf("unsupported media server type: %s", config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Unsupported media server type: %s", config.Global.MediaServer.Type),
				Elapsed: utils.ElapsedTime(startTime),
			},
		}
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logErr)
		return
	}

	// Check if the temporary folder has the image
	fileName := fmt.Sprintf("%s_%s.jpg", ratingKey, imageType)
	filePath := path.Join(tmpFolder, fileName)
	exists := utils.CheckIfImageExists(filePath)
	if exists {
		// Serve the image from the temporary folder
		imagePath := path.Join(tmpFolder, fileName)
		http.ServeFile(w, r, imagePath)
		return
	}

	// If the image does not exist, then get it from the media server
	imageData, logErr := mediaServer.FetchImageFromMediaServer(ratingKey, imageType)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	// If the user has enabled caching, then save the image to the temporary folder
	if config.Global.CacheImages {
		logErr = utils.CheckFolderExists(tmpFolder)
		if logErr.Err != nil {
			utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
			return
		}
		imagePath := path.Join(tmpFolder, fileName)
		err := os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
				Err: err,
				Log: logging.Log{
					Message: "Failed to write image to temporary folder",
					Elapsed: utils.ElapsedTime(startTime),
				},
			})
			return
		}
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "image/jpeg")
	// Write the image data to the response
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}
