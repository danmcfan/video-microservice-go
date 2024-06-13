package internal

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/danmcfan/video-microservice-golang/internal/database"
)

type VideoJob struct {
	ID       string
	Filepath string
}

type VideoInfo struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	RFrameRate string  `json:"r_frame_rate"`
	FrameRate  float64 `json:"-"`
	FrameCount string  `json:"nb_frames"`
}

func RunVideoWorker(db *sql.DB, storageDir string, videoJobs <-chan VideoJob) {
	for job := range videoJobs {
		stream, err := getStream(job.Filepath)
		if err != nil {
			log.Fatalf("Failed to get video stream: %v", err)
			return
		}

		frameCountInt, err := strconv.Atoi(stream.FrameCount)
		if err != nil {
			log.Fatalf("Failed to convert frame count to int: %v", err)
			return
		}

		video := &database.Video{ID: job.ID, Filepath: job.Filepath, Width: stream.Width, Height: stream.Height, FrameRate: stream.FrameRate, FrameCount: frameCountInt}
		err = database.InsertVideo(db, video)
		if err != nil {
			log.Fatalf("Failed to insert video: %v", err)
			return
		}

		framesDir := fmt.Sprintf("%s/frames/%s", storageDir, job.ID)
		err = os.Mkdir(framesDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create frames directory: %v", err)
			return
		}

		err = extractFrames(job.Filepath, framesDir)
		if err != nil {
			log.Fatalf("Failed to extract frames: %v", err)
			return
		}
	}
}

func getStream(filePath string) (*Stream, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var info VideoInfo
	err = json.Unmarshal(out.Bytes(), &info)
	if err != nil {
		return nil, err
	}

	// Calculate the frame rate (FPS)
	for i, stream := range info.Streams {
		rateParts := strings.Split(stream.RFrameRate, "/")
		if len(rateParts) == 2 {
			var num, denom float64
			fmt.Sscanf(rateParts[0], "%f", &num)
			fmt.Sscanf(rateParts[1], "%f", &denom)
			info.Streams[i].FrameRate = num / denom
		}
	}

	return &info.Streams[0], nil
}

func extractFrames(filePath string, outputDir string) error {
	cmd := exec.Command("ffmpeg", "-i", filePath, fmt.Sprintf("%s/%%d.png", outputDir))
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("Command failed with error: %v", err)
		log.Printf("Command output: %s", stderr.String())
		return err
	}
	return nil
}
