package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RenameResponse represents the response from the rename API
type RenameResponse struct {
	Errno     int          `json:"errno"`
	Info      []RenameInfo `json:"info"`
	TaskID    int64        `json:"taskid"`
	RequestID int64        `json:"request_id"`
}

// RenameInfo represents the result for each renamed file in the response
type RenameInfo struct {
	Path  string `json:"path"`
	Errno int    `json:"errno"`
}

// RenameRequest represents the structure for a file to be renamed
type RenameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"newname"`
}

// RenameFile renames a single file or directory in Baidu Pan
func (c *Client) RenameFile(sourcePath, newName string) error {
	renameRequests := []RenameRequest{
		{
			Path:    sourcePath,
			NewName: newName,
		},
	}
	return c.RenameFiles(renameRequests)
}

// RenameFiles renames multiple files based on the provided RenameRequest structs
func (c *Client) RenameFiles(renameRequests []RenameRequest) error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token, please authorize first")
	}

	if len(renameRequests) == 0 {
		return fmt.Errorf("no files specified for rename operation")
	}

	// Convert rename requests to JSON format for the POST body
	renameRequestsJSON, err := json.Marshal(renameRequests)
	if err != nil {
		return fmt.Errorf("failed to marshal rename requests to JSON: %w", err)
	}

	// Prepare query parameters
	params := url.Values{}
	params.Add("method", "filemanager")
	params.Add("access_token", c.accessToken)
	params.Add("opera", "rename")

	// Additional parameters that might be required based on API documentation
	params.Add("channel", "chunlei")
	params.Add("web", "1")
	params.Add("app_id", "250528")
	params.Add("bdstoken", c.accessToken) // Using access token as bdstoken (common practice)

	// Create form data for POST body
	formData := url.Values{}
	formData.Add("filelist", string(renameRequestsJSON))
	// Optional: Add ondup parameter to handle duplicate files (default is "fail")
	formData.Add("ondup", "newcopy")
	// Use synchronous operation
	formData.Add("async", "0")

	// Create the request with form-encoded body
	apiURL := fmt.Sprintf("https://pan.baidu.com/api/filemanager?%s", params.Encode())
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create rename request: %w", err)
	}

	// Set content type for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("rename request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read rename response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rename request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the response
	var renameResponse RenameResponse
	err = json.Unmarshal(responseBody, &renameResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal rename response: %w", err)
	}

	if renameResponse.Errno != 0 {
		return fmt.Errorf("rename API returned error code %d: %s", renameResponse.Errno, GetRenameErrorMessage(renameResponse.Errno))
	}

	// Check if any individual files failed to rename
	var failedRenames []string
	for i, renameInfo := range renameResponse.Info {
		if renameInfo.Errno != 0 {
			req := renameRequests[i] // Get the corresponding request
			failedRenames = append(failedRenames, fmt.Sprintf("%s -> %s (error code: %d)", req.Path, req.NewName, renameInfo.Errno))
		}
	}

	if len(failedRenames) > 0 {
		return fmt.Errorf("failed to rename some files: %s", strings.Join(failedRenames, "; "))
	}

	return nil
}

// GetRenameErrorMessage returns a human-readable error message for common errno values
func GetRenameErrorMessage(errno int) string {
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
