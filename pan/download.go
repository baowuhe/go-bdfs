package pan

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// DownloadFile downloads a file from Baidu Pan
func (c *Client) DownloadFile(filePath string) (*http.Response, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("no access token, please authorize first")
	}

	params := url.Values{}
	params.Add("method", "download")
	params.Add("access_token", c.accessToken)
	params.Add("path", filePath)

	req, err := http.NewRequest("GET", downloadFileURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	return c.downloadClient.Do(req) // Use downloadClient with longer timeout
}

// ProgressWriter wraps an io.Writer and reports progress
type ProgressWriter struct {
	writer     io.Writer
	totalSize  int64
	downloaded int64
	fileName   string
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.downloaded += int64(n)

	// Calculate percentage
	var percent float64
	if pw.totalSize > 0 {
		percent = float64(pw.downloaded) / float64(pw.totalSize) * 100
	}

	// Print progress on the same line
	fmt.Printf("\rDownloading %s: %d / %d bytes (%.2f%%)",
		pw.fileName, pw.downloaded, pw.totalSize, percent)
	os.Stdout.Sync()

	return n, err
}

// DownloadFileToPath downloads a file from Baidu Pan and saves it to the specified local path
func (c *Client) DownloadFileToPath(filePath, localPath string) error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token, please authorize first")
	}

	// Check if the directory for the local path exists, create if not
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for local path: %w", err)
	}

	// Get file information to know the total size
	fileInfo, err := c.GetFileInfo(filePath)
	if err != nil {
		// If we can't get file info, proceed with download anyway but without size info
		PrintError(fmt.Sprintf("Warning: Could not get file size information: %v", err))
	}

	// Download the file content
	resp, err := c.DownloadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to download file from Baidu Pan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Create the local file
	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer outFile.Close()

	// Create progress writer if we have file info
	var writer io.Writer
	if fileInfo != nil {
		progressWriter := &ProgressWriter{
			writer:     outFile,
			totalSize:  fileInfo.Size,
			downloaded: 0,
			fileName:   fileInfo.ServerFilename,
		}
		writer = progressWriter
	} else {
		// If we don't have file info, just use the outFile directly
		_, fileName := filepath.Split(filePath)
		fmt.Printf("Downloading %s...\n", fileName)
		writer = outFile
	}

	// Copy the response body to the local file with progress reporting
	buf := make([]byte, 32*1024) // 32KB buffer
	_, err = io.CopyBuffer(writer, resp.Body, buf)
	if err != nil {
		// Clean up the partially downloaded file if there's an error
		os.Remove(localPath)
		return fmt.Errorf("failed to write file content to local file: %w", err)
	}

	// Print final progress and newline
	if fileInfo != nil {
		fmt.Printf("\n") // Newline after progress is complete
	}

	return nil
}

// ReadFileContent reads the content of a file from Baidu Pan
func (c *Client) ReadFileContent(filePath string) ([]byte, error) {
	resp, err := c.DownloadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}
