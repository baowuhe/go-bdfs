package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GetFileInfoByPath gets information about a specific file or directory by its path
// This method uses the list API with a filter to get information about a single file
func (c *Client) GetFileInfoByPath(filePath string) (*FileInfo, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("no access token, please authorize first")
	}

	// Use the list API with filename filter to get specific file info
	dirPath := "/"
	if filePath != "/" {
		// Extract parent directory
		lastSlash := strings.LastIndex(filePath, "/")
		if lastSlash > 0 {
			dirPath = filePath[:lastSlash]
		} else if lastSlash == 0 {
			dirPath = "/"
		}
	}

	filename := filePath
	if lastSlash := strings.LastIndex(filePath, "/"); lastSlash >= 0 && lastSlash < len(filePath)-1 {
		filename = filePath[lastSlash+1:]
	}

	params := url.Values{}
	params.Add("method", "list")
	params.Add("access_token", c.accessToken)
	params.Add("dir", dirPath)
	params.Add("filename", filename) // Filter by filename
	params.Add("folder", "0")

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
		return nil, fmt.Errorf("get file info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ListFilesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Errno != 0 {
		return nil, fmt.Errorf("API returned error code %d", response.Errno)
	}

	// Find the file with matching path
	for _, file := range response.List {
		if file.Path == filePath {
			return &file, nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", filePath)
}

// GetDetailedFileInfo gets detailed information about a file using the meta API
// This is more efficient than listing files when you only need info about one file
func (c *Client) GetDetailedFileInfo(filePath string) (*FileInfo, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("no access token, please authorize first")
	}

	// Use the meta API to get detailed information about a single file
	params := url.Values{}
	params.Add("method", "meta")
	params.Add("access_token", c.accessToken)
	params.Add("path", filePath)

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
		return nil, fmt.Errorf("get file meta request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var metaResponse struct {
		Errno int        `json:"errno"`
		List  []FileInfo `json:"list"`
	}

	err = json.Unmarshal(body, &metaResponse)
	if err != nil {
		return nil, err
	}

	if metaResponse.Errno != 0 {
		return nil, fmt.Errorf("API returned error code %d", metaResponse.Errno)
	}

	if len(metaResponse.List) == 0 {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	return &metaResponse.List[0], nil
}

// GetAndDisplayFileInfo gets file information from Baidu Pan and returns formatted information
// This encapsulates the logic for getting file info from Baidu Pan and processing it
func (c *Client) GetAndDisplayFileInfo(filePath string) (*FileInfo, error) {
	// Try to get detailed file info using the meta API (more efficient)
	fileInfo, err := c.GetDetailedFileInfo(filePath)
	if err != nil {
		// Fallback to the original GetFileInfo method if meta API fails
		fileInfo, err = c.GetFileInfo(filePath) // Use existing method from list.go
		if err != nil {
			return nil, fmt.Errorf("error getting file information: %w", err)
		}
	}

	return fileInfo, nil
}

// FormatFileInfo formats the file information in a human-readable way
func FormatFileInfo(fileInfo *FileInfo) string {
	var result strings.Builder

	result.WriteString("File Information:\n")
	result.WriteString(fmt.Sprintf("  Name: %s\n", fileInfo.ServerFilename))
	result.WriteString(fmt.Sprintf("  Path: %s\n", fileInfo.Path))
	result.WriteString(fmt.Sprintf("  Size: %d bytes\n", fileInfo.Size))
	result.WriteString(fmt.Sprintf("  Type: %s\n", MapFileType(fileInfo.IsDir)))
	result.WriteString(fmt.Sprintf("  MD5: %s\n", fileInfo.MD5))
	result.WriteString(fmt.Sprintf("  File ID: %d\n", fileInfo.FsID))
	result.WriteString(fmt.Sprintf("  Created: %s\n", FormatTime(fileInfo.ServerCtime)))
	result.WriteString(fmt.Sprintf("  Modified: %s\n", FormatTime(fileInfo.ServerMtime)))
	result.WriteString(fmt.Sprintf("  Category: %d\n", fileInfo.Category))
	result.WriteString(fmt.Sprintf("  Real Category: %s\n", fileInfo.RealCategory))

	return result.String()
}

// MapFileType converts the isdir field to a readable file type
func MapFileType(isDir int) string {
	if isDir == 1 {
		return "Directory"
	}
	return "File"
}

// FormatTime converts Unix timestamp to readable time format
func FormatTime(unixTime int64) string {
	if unixTime == 0 {
		return "N/A"
	}
	return time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
}
