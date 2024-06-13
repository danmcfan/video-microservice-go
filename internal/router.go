package internal

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/danmcfan/video-microservice-golang/internal/database"
)

func RunServer(port int, db *sql.DB, storageDir string) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := createRouter(db, storageDir)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	stop()
	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func createRouter(db *sql.DB, storageDir string) *gin.Engine {
	router := gin.Default()

	err := os.MkdirAll(filepath.Join(storageDir, "videos"), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create videos directory: %v", err)
	}

	err = os.MkdirAll(filepath.Join(storageDir, "frames"), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create frames directory: %v", err)
	}

	videoJobs := make(chan VideoJob, 100)

	for i := 0; i < 10; i++ {
		go RunVideoWorker(db, storageDir, videoJobs)
	}

	router.POST("/videos", func(c *gin.Context) {
		formFile, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No video file found"})
			return
		}

		id := uuid.New().String()
		videoFilePath := filepath.Join(storageDir, "videos", id+".mp4")
		if err := c.SaveUploadedFile(formFile, videoFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video file"})
			return
		}

		videoJobs <- VideoJob{ID: id, Filepath: videoFilePath}

		c.JSON(http.StatusOK, gin.H{"id": id})
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

		frame, err := database.GetFrame(db, id, index, storageDir)
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

		frame, err := database.GetFrame(db, id, index, storageDir)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Frame not found"})
			return
		}

		c.File(frame.Filepath)
	})

	return router
}
