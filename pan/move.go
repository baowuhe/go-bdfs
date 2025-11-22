package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// MoveResponse represents the response from the move API
type MoveResponse struct {
	Errno     int        `json:"errno"`
	Info      []MoveInfo `json:"info"`
	TaskID    int64      `json:"taskid"`
	RequestID int64      `json:"request_id"`
}

// MoveInfo represents the result for each moved file in the response
type MoveInfo struct {
	Path    string `json:"path"`
	Dest    string `json:"dest"`
	NewName string `json:"newname"`
	Errno   int    `json:"errno"`
}

// MoveRequest represents the structure for a file to be moved
type MoveRequest struct {
	Path    string `json:"path"`
	Dest    string `json:"dest"`
	NewName string `json:"newname"`
}

// MoveFile moves a single file or directory from source path to destination directory in Baidu Pan
func (c *Client) MoveFile(sourcePath, destDir string) error {
	// Ensure paths are in the correct format
	sourcePath = strings.TrimRight(sourcePath, "/")
	destDir = strings.TrimRight(destDir, "/")

	moveRequests := []MoveRequest{
		{
			Path:    sourcePath,
			Dest:    destDir,
			NewName: GetSourceFileName(sourcePath),
		},
	}
	return c.MoveFiles(moveRequests)
}

// MoveFiles moves multiple files based on the provided MoveRequest structs
func (c *Client) MoveFiles(moveRequests []MoveRequest) error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token, please authorize first")
	}

	if len(moveRequests) == 0 {
		return fmt.Errorf("no files specified for move operation")
	}

	// We'll attempt the move operation directly since the API handles both files and directories
	// Path validation is handled by the API itself

	// Convert move requests to JSON format for the POST body
	moveRequestsJSON, err := json.Marshal(moveRequests)
	if err != nil {
		return fmt.Errorf("failed to marshal move requests to JSON: %w", err)
	}

	// Prepare query parameters
	params := url.Values{}
	params.Add("method", "filemanager")
	params.Add("access_token", c.accessToken)
	params.Add("opera", "move")

	// Additional parameters that might be required based on API documentation
	params.Add("async", "0") // synchronous operation
	params.Add("channel", "chunlei")
	params.Add("web", "1")
	params.Add("app_id", "250528")
	params.Add("bdstoken", c.accessToken) // Using access token as bdstoken (common practice)

	// Create form data for POST body
	formData := url.Values{}
	formData.Add("filelist", string(moveRequestsJSON))
	// Optional: Add ondup parameter to handle duplicate files (default is "fail")
	formData.Add("ondup", "newcopy")

	// Create the request with form-encoded body
	apiURL := fmt.Sprintf("https://pan.baidu.com/api/filemanager?%s", params.Encode())
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create move request: %w", err)
	}

	// Set content type for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("move request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read move response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("move request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the response
	var moveResponse MoveResponse
	err = json.Unmarshal(responseBody, &moveResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal move response: %w", err)
	}

	if moveResponse.Errno != 0 {
		return fmt.Errorf("move API returned error code %d: %s", moveResponse.Errno, GetMoveErrorMessage(moveResponse.Errno))
	}

	// Check if any individual files failed to move
	var failedMoves []string
	for i, moveInfo := range moveResponse.Info {
		if moveInfo.Errno != 0 {
			req := moveRequests[i] // Get the corresponding request
			failedMoves = append(failedMoves, fmt.Sprintf("%s -> %s/%s (error code: %d)", req.Path, req.Dest, req.NewName, moveInfo.Errno))
		}
	}

	if len(failedMoves) > 0 {
		return fmt.Errorf("failed to move some files: %s", strings.Join(failedMoves, "; "))
	}

	return nil
}

// GetMoveErrorMessage returns a human-readable error message for common errno values
func GetMoveErrorMessage(errno int) string {
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
