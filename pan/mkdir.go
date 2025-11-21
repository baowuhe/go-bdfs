package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CreateDir creates a directory in Baidu Pan
func (c *Client) CreateDir(remotePath string) error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token, please authorize first")
	}

	// Ensure remote path is valid
	if !strings.HasPrefix(remotePath, "/") {
		return fmt.Errorf("remote directory path must be an absolute path starting with '/'")
	}

	// Prepare parameters for the create API
	params := url.Values{}
	params.Add("method", "create")
	params.Add("access_token", c.accessToken)
	params.Add("path", remotePath)
	params.Add("isdir", "1")       // 1 for directory, 0 for file
	params.Add("block_list", "[]") // Empty block list for directories

	// Create the POST request
	req, err := http.NewRequest("POST", uploadCreateFileUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create directory creation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("directory creation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read directory creation response: %w", err)
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("directory creation API failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response to check for API-specific errors
	var response CreateFileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal directory creation response: %w", err)
	}

	// Check if the API returned an error code
	if response.Errno != 0 {
		return fmt.Errorf("directory creation API returned error code %d. Response: %s", response.Errno, string(body))
	}

	// Success
	PrintSuccess(fmt.Sprintf("Directory '%s' created successfully in Baidu Pan.", remotePath))
	return nil
}
