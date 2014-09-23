package main

import (
	"os/exec"
	"os"
)

func StartJobs(srcPath, finalPath string) error {
	args := []string{
		"-i", srcPath, "-map", "0:0", "-map", "0:1", "-c:v", "copy", "-c:a", "copy", finalPath,
	}

	cmd := exec.Command("avconv", args...)
	if err := cmd.Run(); err != nil {
		os.Remove(finalPath)
		return err
	}

	os.Remove(srcPath)

	return nil
}
