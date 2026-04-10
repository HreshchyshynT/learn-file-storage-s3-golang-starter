package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

const tolerance = 0.1

func GetVideoAspectRatio(filePath string) (string, error) {
	type probeResult struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	var result probeResult

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("unable to get video aspect ratio: %v", err.Error())
	}

	err = json.Unmarshal(output, &result)

	if err != nil {
		return "", fmt.Errorf("unable to decode video information: %v", err.Error())
	}

	var aspectRatio string

	for _, stream := range result.Streams {
		if stream.Width > 0 && stream.Height > 0 {
			ar := float64(stream.Width) / float64(stream.Height)
			switch {
			case math.Abs(16.0/9.0-ar) < tolerance:
				aspectRatio = "16:9"
			case math.Abs(9.0/16.0-ar) < tolerance:
				aspectRatio = "9:16"
			default:
				aspectRatio = "other"
			}
			break
		}
	}

	return aspectRatio, nil
}
