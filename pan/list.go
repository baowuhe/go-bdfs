package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ListFiles lists files in a directory
func (c *Client) ListFiles(dirPath string) ([]FileInfo, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("no access token, please authorize first")
	}

	params := url.Values{}
	params.Add("method", "list")
	params.Add("access_token", c.accessToken)
	params.Add("dir", dirPath)
	params.Add("folder", "0") // 0 for all files, 1 for folders only

	req, err := http.NewRequest("GET", listFilesURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list files request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ListFilesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Errno != 0 {
		return nil, fmt.Errorf("API returned error code %d", response.Errno)
	}

	return response.List, nil
}

// GetFileInfo gets information about a specific file
func (c *Client) GetFileInfo(filePath string) (*FileInfo, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("no access token, please authorize first")
	}

	// List files in the parent directory and find our file
	dirPath := "/"
	if strings.Contains(filePath, "/") {
		lastSlash := strings.LastIndex(filePath, "/")
		dirPath = filePath[:lastSlash]
		if dirPath == "" {
			dirPath = "/"
		}
	}

	files, err := c.ListFiles(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Path == filePath {
			return &file, nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", filePath)
}

// Walk recursively walks through directories and files
func (c *Client) Walk(rootPath string) (<-chan FileInfo, <-chan error) {
	fileChan := make(chan FileInfo)
	errChan := make(chan error, 1)

	go func() {
		defer close(fileChan)
		c.WalkRecursive(rootPath, fileChan, errChan)
	}()

	return fileChan, errChan
}

func (c *Client) WalkRecursive(path string, fileChan chan<- FileInfo, errChan chan<- error) {
	files, err := c.ListFiles(path)
	if err != nil {
		errChan <- err
		return
	}

	for _, file := range files {
		fileChan <- file

		// If it's a directory, recurse into it
		if file.IsDir == 1 {
			subPath := file.Path
			c.WalkRecursive(subPath, fileChan, errChan)
		}
	}
}
