package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DeleteResponse represents the response from the delete API
type DeleteResponse struct {
	Errno     int           `json:"errno"`
	RequestID int64         `json:"request_id"`
	List      []DeleteEntry `json:"list"`
}

// DeleteEntry represents the result for each file in the delete operation
type DeleteEntry struct {
	Path  string `json:"path"`
	Errno int    `json:"errno"`
}

// RemoveFile removes a single file or directory from Baidu Pan
func (c *Client) RemoveFile(filePath string) error {
	return c.RemoveFiles([]string{filePath})
}

// RemoveFiles removes multiple files or directories from Baidu Pan
func (c *Client) RemoveFiles(filePaths []string) error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token, please authorize first")
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("no files specified for deletion")
	}

	// Convert file paths to JSON format for the POST body
	fileListJSON := "["
	for i, path := range filePaths {
		if i > 0 {
			fileListJSON += ","
		}
		fileListJSON += fmt.Sprintf("\"%s\"", path)
	}
	fileListJSON += "]"

	// Prepare query parameters - using the correct endpoint according to documentation
	params := url.Values{}
	params.Add("method", "filemanager")
	params.Add("access_token", c.accessToken)
	params.Add("opera", "delete")

	// Additional parameters that might be required based on API documentation
	params.Add("async", "0") // synchronous operation
	params.Add("channel", "chunlei")
	params.Add("web", "1")
	params.Add("app_id", "250528")
	params.Add("bdstoken", c.accessToken) // Using access token as bdstoken (common practice)

	// Create form data for POST body
	formData := url.Values{}
	formData.Add("filelist", fileListJSON)

	// Create the request with form-encoded body
	req, err := http.NewRequest("POST", "https://pan.baidu.com/api/filemanager?"+params.Encode(), strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	// Set content type for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read delete response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the response
	var deleteResponse DeleteResponse
	err = json.Unmarshal(responseBody, &deleteResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal delete response: %w", err)
	}

	if deleteResponse.Errno != 0 {
		return fmt.Errorf("delete API returned error code %d: %s", deleteResponse.Errno, GetErrorMessage(deleteResponse.Errno))
	}

	// Check if any individual files failed to delete
	var failedFiles []string
	for _, entry := range deleteResponse.List {
		if entry.Errno != 0 {
			failedFiles = append(failedFiles, fmt.Sprintf("%s (error code: %d)", entry.Path, entry.Errno))
		}
	}

	if len(failedFiles) > 0 {
		return fmt.Errorf("failed to delete some files: %s", strings.Join(failedFiles, "; "))
	}

	return nil
}

// GetErrorMessage returns a human-readable error message for common errno values
func GetErrorMessage(errno int) string {
	switch errno {
	case 0:
		return "Success"
	case 2:
		return "Parameters error"
	case 3:
		return "User permission error"
	case 4:
		return "Request source error"
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
