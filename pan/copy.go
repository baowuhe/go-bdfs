package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// CopyResponse represents the response from the copy API
type CopyResponse struct {
	Errno     int        `json:"errno"`
	Info      []CopyInfo `json:"info"`
	TaskID    int64      `json:"taskid"`
	RequestID int64      `json:"request_id"`
}

// CopyInfo represents the result for each copied file in the response
type CopyInfo struct {
	Path    string `json:"path"`
	Dest    string `json:"dest"`
	NewName string `json:"newname"`
	Errno   int    `json:"errno"`
}

// CopyRequest represents the structure for a file to be copied
type CopyRequest struct {
	Path    string `json:"path"`
	Dest    string `json:"dest"`
	NewName string `json:"newname"`
}

// CopyFile copies a single file or directory from source path to destination directory in Baidu Pan
func (c *Client) CopyFile(sourcePath, destPath string) error {
	// Ensure paths are in the correct format
	sourcePath = strings.TrimRight(sourcePath, "/")
	destPath = strings.TrimRight(destPath, "/")

	// Extract destination directory and new filename
	destDir := filepath.Dir(destPath)
	newName := filepath.Base(destPath)

	// If destPath is a directory, use the source filename
	if strings.HasSuffix(destPath, "/") || isDirectoryPath(destPath) {
		destDir = destPath
		newName = GetSourceFileName(sourcePath)
	}

	copyRequests := []CopyRequest{
		{
			Path:    sourcePath,
			Dest:    destDir,
			NewName: newName,
		},
	}
	return c.CopyFiles(copyRequests)
}

// CopyFiles copies multiple files based on the provided CopyRequest structs
func (c *Client) CopyFiles(copyRequests []CopyRequest) error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token, please authorize first")
	}

	if len(copyRequests) == 0 {
		return fmt.Errorf("no files specified for copy operation")
	}

	// Convert copy requests to JSON format for the POST body
	copyRequestsJSON, err := json.Marshal(copyRequests)
	if err != nil {
		return fmt.Errorf("failed to marshal copy requests to JSON: %w", err)
	}

	// Prepare query parameters
	params := url.Values{}
	params.Add("method", "filemanager")
	params.Add("access_token", c.accessToken)
	params.Add("opera", "copy")

	// Additional parameters that might be required based on API documentation
	params.Add("async", "0") // synchronous operation
	params.Add("channel", "chunlei")
	params.Add("web", "1")
	params.Add("app_id", "250528")
	params.Add("bdstoken", c.accessToken) // Using access token as bdstoken (common practice)

	// Create form data for POST body
	formData := url.Values{}
	formData.Add("filelist", string(copyRequestsJSON))
	// Optional: Add ondup parameter to handle duplicate files (default is "fail")
	formData.Add("ondup", "newcopy")

	// Create the request with form-encoded body
	apiURL := fmt.Sprintf("https://pan.baidu.com/api/filemanager?%s", params.Encode())
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create copy request: %w", err)
	}

	// Set content type for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("copy request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read copy response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("copy request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the response
	var copyResponse CopyResponse
	err = json.Unmarshal(responseBody, &copyResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal copy response: %w", err)
	}

	if copyResponse.Errno != 0 {
		return fmt.Errorf("copy API returned error code %d: %s", copyResponse.Errno, GetCopyErrorMessage(copyResponse.Errno))
	}

	// Check if any individual files failed to copy
	var failedCopies []string
	for i, copyInfo := range copyResponse.Info {
		if copyInfo.Errno != 0 {
			req := copyRequests[i] // Get the corresponding request
			failedCopies = append(failedCopies, fmt.Sprintf("%s -> %s/%s (error code: %d)", req.Path, req.Dest, req.NewName, copyInfo.Errno))
		}
	}

	if len(failedCopies) > 0 {
		return fmt.Errorf("failed to copy some files: %s", strings.Join(failedCopies, "; "))
	}

	// Print success message for each successfully copied file
	for i, _ := range copyResponse.Info {
		req := copyRequests[i]
		PrintSuccess(fmt.Sprintf("File '%s' copied successfully to '%s/%s'", req.Path, req.Dest, req.NewName))
	}

	return nil
}

// isDirectoryPath checks if the given path likely refers to a directory
func isDirectoryPath(path string) bool {
	// If the path ends with "/" or doesn't have a file extension, it's likely a directory
	return strings.HasSuffix(path, "/") || filepath.Ext(path) == ""
}

// GetCopyErrorMessage returns a human-readable error message for common errno values
func GetCopyErrorMessage(errno int) string {
	switch errno {
	case 0:
		return "Success"
	case 2:
		return "Parameters error"
	case 3:
		return "User permission error"
	case 4:
		return "Request source error"
	case 12:
		return "Operation not allowed or path error"
	case -9:
		return "File does not exist"
	case 111:
		return "Another asynchronous task is currently executing"
	case -7:
		return "Invalid file name"
	case 108:
		return "Path error, path does not exist"
	case 110:
		return "Target path already exists"
	case 112:
		return "Same file already exists in the same directory"
	case 113:
		return "File or directory name contains forbidden words"
	case 114:
		return "Path too long"
	case 115:
		return "Target directory does not exist"
	case 116:
		return "Insufficient disk space"
	case 117:
		return "File too large"
	case 31001:
		return "User has been banned"
	case 31026:
		return "File contains illegal content"
	default:
		return fmt.Sprintf("Unknown error code: %d", errno)
	}
}
