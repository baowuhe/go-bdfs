package pan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DiskInfoResponse represents the response from the disk info API
type DiskInfoResponse struct {
	Errno     int64  `json:"errno"`
	Total     int64  `json:"total"`     // Total space size in bytes
	Used      int64  `json:"used"`      // Used space size in bytes
	Free      int64  `json:"free"`      // Free capacity in bytes
	Expire    bool   `json:"expire"`    // Whether capacity will expire within 7 days
	RequestID int64  `json:"request_id"`
}

// GetDiskInfo gets the user's cloud storage usage information
func (c *Client) GetDiskInfo() (*DiskInfoResponse, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("no access token, please authorize first")
	}

	params := url.Values{}
	params.Add("access_token", c.accessToken)
	params.Add("checkfree", "1")      // Check free information
	params.Add("checkexpire", "1")    // Check expiration information

	// Baidu Pan quota API endpoint
	apiURL := "https://pan.baidu.com/api/quota"

	req, err := http.NewRequest("GET", apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent as required by Baidu Pan API
	req.Header.Set("User-Agent", "pan.baidu.com")

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
		return nil, fmt.Errorf("get disk info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response DiskInfoResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Errno != 0 {
		return nil, fmt.Errorf("API returned error code %d", response.Errno)
	}

	return &response, nil
}

// FormatDiskInfo formats the disk information in a human-readable way
func FormatDiskInfo(info *DiskInfoResponse) string {
	var result string

	result += "Disk Information:\n"
	result += fmt.Sprintf("  Total: %s\n", FormatBytes(info.Total))
	result += fmt.Sprintf("  Used:  %s\n", FormatBytes(info.Used))
	result += fmt.Sprintf("  Free:  %s\n", FormatBytes(info.Free))
	result += fmt.Sprintf("  Expire: %t\n", info.Expire)

	return result
}

// FormatBytes converts bytes to a human-readable format (e.g., KB, MB, GB)
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}