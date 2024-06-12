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
}

func RunVideoWorker(db *sql.DB, storageDir string, videoJobs <-chan VideoJob) {
	for job := range videoJobs {
		stream, err := GetVideoStream(job.Filepath)
		if err != nil {
			log.Fatalf("Failed to get video stream: %v", err)
			return
		}

		frameCount, err := GetFrameCount(job.Filepath)
		if err != nil {
			log.Fatalf("Failed to get frame count: %v", err)
			return
		}

		video := &database.Video{ID: job.ID, Filepath: job.Filepath, Width: stream.Width, Height: stream.Height, FrameRate: stream.FrameRate, FrameCount: frameCount}
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

		err = ExtractFrames(job.Filepath, framesDir, int(stream.FrameRate))
		if err != nil {
			log.Fatalf("Failed to extract frames: %v", err)
			return
		}
	}
}

func GetVideoStream(filePath string) (*Stream, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height,r_frame_rate", "-of", "json", filePath)
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

func GetFrameCount(filePath string) (int, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-count_frames", "-show_entries", "stream=nb_read_frames", "-of", "json", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var result struct {
		Streams []struct {
			FrameCount string `json:"nb_read_frames"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	if len(result.Streams) == 0 {
		return 0, fmt.Errorf("no video stream found in file")
	}

	frameCount, err := strconv.Atoi(strings.TrimSpace(result.Streams[0].FrameCount))
	if err != nil {
		return 0, fmt.Errorf("failed to convert frame count to integer: %w", err)
	}

	return frameCount, nil
}

func ExtractFrames(filePath string, outputDir string, fps int) error {
	cmd := exec.Command("ffmpeg", "-i", filePath, "-q:v", "2", "-vf", "fps="+strconv.Itoa(fps), "-vsync", "vfr", fmt.Sprintf("%s/%%d.png", outputDir))
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
