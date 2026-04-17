package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func ProcessVideoForFastStart(filePath string) (string, error) {
	base := filepath.Base(filePath)
	if base == "." {
		return "", errors.New("invalid file path")
	}

	splitted := strings.Split(base, ".")
	output := fmt.Sprint(splitted[0], ".processing", ".", splitted[1])

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", output)
	_, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return output, nil
}
