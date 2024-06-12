package database

import (
	"database/sql"
	"fmt"
	"os"
)

type Frame struct {
	VideoID     string `json:"video_id"`
	FrameNumber int    `json:"frame_number"`
	Filepath    string `json:"-"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func GetFrame(db *sql.DB, videoID string, frameNumber int, storageDir string) (*Frame, error) {
	video, err := GetVideo(db, videoID)
	if err != nil {
		return nil, err
	}

	frameFilePath := fmt.Sprintf("%s/frames/%s/%d.png", storageDir, videoID, frameNumber)

	_, err = os.Stat(frameFilePath)
	frameExists := !os.IsNotExist(err)
	if !frameExists {
		return nil, fmt.Errorf("frame %d does not exist", frameNumber)
	}

	return &Frame{
		VideoID:     videoID,
		FrameNumber: frameNumber,
		Filepath:    frameFilePath,
		Width:       video.Width,
		Height:      video.Height,
	}, nil
}
