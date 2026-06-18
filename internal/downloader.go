package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ProgressTracker struct {
	TotalBytes   int64
	BytesWritten int64
}

func (pt *ProgressTracker) Write(p []byte) (int, error) {
	n := len(p)
	pt.BytesWritten += int64(n)

	if pt.TotalBytes > 0 {
		percentage := (float64(pt.BytesWritten) / float64(pt.TotalBytes)) * 100
		fmt.Printf("\rDownloading... %.2f%% complete", percentage)

		// 🚀 WAILS FUTURE: This is where you will add:
		// runtime.EventsEmit(pt.ctx, "download:progress", percentage)
	}

	return n, nil
}

func DownloadModfile(url string, paths AppPaths) (string, error) {
	filename := filepath.Base(url)
	if filename == "." || filename == "/" {
		return "", fmt.Errorf("invalid or empty filename extracted from URL")
	}

	destPath := filepath.Join(paths.RawDownloadsDir, filename)

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to connect to download URL: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned bad status: %s", resp.Status)
	}

	totalBytes := resp.ContentLength

	tracker := &ProgressTracker{
		TotalBytes: totalBytes,
	}

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file on disk: %w", err)
	}

	defer out.Close()

	ProgressReader := io.TeeReader(resp.Body, tracker)

	_, err = io.Copy(out, ProgressReader)
	if err != nil {
		return "", fmt.Errorf("download interrupted while streaming data: %w", err)
	}

	return destPath, nil
}
