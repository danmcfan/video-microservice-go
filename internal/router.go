package internal

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/danmcfan/video-microservice-golang/internal/database"
)

func CreateRouter(db *sql.DB, storageDir string) *gin.Engine {
	router := gin.Default()

	router.POST("/videos", func(c *gin.Context) {
		// Get the video file
		formFile, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No video file found"})
			return
		}

		id := uuid.New().String()
		videoFilePath := filepath.Join(storageDir, id+".mp4")
		if err := c.SaveUploadedFile(formFile, videoFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video file"})
			return
		}

		stream, err := GetVideoStream(videoFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video stream info"})
			return
		}

		frameCount, err := GetFrameCount(videoFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video frame count: " + err.Error()})
			return
		}

		video := &database.Video{ID: id, Filepath: videoFilePath, Width: stream.Width, Height: stream.Height, FrameRate: stream.FrameRate, FrameCount: frameCount}
		err = database.InsertVideo(db, video)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert video into database"})
			return
		}

		c.JSON(http.StatusOK, video)
	})

	router.GET("/videos", func(c *gin.Context) {
		videos, err := database.ListVideos(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
			return
		}

		c.JSON(http.StatusOK, videos)
	})

	router.GET("/videos/:id", func(c *gin.Context) {
		id := c.Param("id")
		video, err := database.GetVideo(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		c.JSON(http.StatusOK, video)
	})

	router.GET("/videos/:id/content", func(c *gin.Context) {
		id := c.Param("id")
		video, err := database.GetVideo(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		c.File(video.Filepath)
	})

	router.GET("/videos/:id/frames/:index", func(c *gin.Context) {
		id, indexParam := c.Param("id"), c.Param("index")
		index, err := strconv.Atoi(indexParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid frame index"})
			return
		}

		frame, err := database.GetFrame(db, id, index)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Frame not found"})
			return
		}

		c.JSON(http.StatusOK, frame)
	})

	router.GET("/videos/:id/frames/:index/content", func(c *gin.Context) {
		id, indexParam := c.Param("id"), c.Param("index")
		index, err := strconv.Atoi(indexParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid frame index"})
			return
		}

		frame, err := database.GetFrame(db, id, index)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Frame not found"})
			return
		}

		c.File(frame.Filepath)
	})

	return router
}
