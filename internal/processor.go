package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type VideoInfo struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	RFrameRate string  `json:"r_frame_rate"`
	FrameRate  float64 `json:"-"`
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
