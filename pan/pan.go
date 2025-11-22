package pan

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Baidu Pan API endpoints
	deviceCodeURL       = "https://openapi.baidu.com/oauth/2.0/device/code"
	accessTokenURL      = "https://openapi.baidu.com/oauth/2.0/token"
	listFilesURL        = "https://pan.baidu.com/rest/2.0/xpan/file"
	downloadFileURL     = "https://pan.baidu.com/rest/2.0/xpan/file"
	uploadPrecreateURL  = "https://pan.baidu.com/rest/2.0/xpan/file?method=precreate"
	uploadSuperfileURL  = "https://d.pcs.baidu.com/rest/2.0/pcs/superfile2"
	uploadCreateFileUrl = "https://pan.baidu.com/rest/2.0/xpan/file?method=create"
)

// DeviceCodeResponse represents the response from device code endpoint
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	QRCode          string `json:"qrcode,omitempty"`
}

// TokenResponse represents the access token response
type TokenResponse struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	Scope         string `json:"scope"`
	SessionKey    string `json:"session_key"`
	SessionSecret string `json:"session_secret"`
	UID           string `json:"uid"`
}

// FileInfo represents a file or directory in Baidu Pan
type FileInfo struct {
	TkBindID       int    `json:"tkbind_id"`
	OwnerType      int    `json:"owner_type"`
	Category       int    `json:"category"`
	IsScene        int    `json:"is_scene"`
	FsID           int64  `json:"fs_id"`
	BlackTag       int    `json:"black_tag"`
	ServerFilename string `json:"server_filename"`
	ExtentInt2     int    `json:"extent_int2"`
	ServerAtime    int64  `json:"server_atime"`
	ServerCtime    int64  `json:"server_ctime"`
	ExtentInt8     int    `json:"extent_int8"`
	WpFile         int    `json:"wpfile"`
	Unlist         int    `json:"unlist"`
	LocalMtime     int64  `json:"local_mtime"`
	Size           int64  `json:"size"`
	OperID         int64  `json:"oper_id"`
	OwnerID        int64  `json:"owner_id"`
	Share          int    `json:"share"`
	FromType       int    `json:"from_type"`
	Path           string `json:"path"`
	LocalCTime     int64  `json:"local_ctime"`
	PL             int    `json:"pl"`
	RealCategory   string `json:"real_category"`
	IsDir          int    `json:"isdir"`
	ExtentTinyInt7 int    `json:"extent_tinyint7"`
	ServerMtime    int64  `json:"server_mtime"`
	MD5            string `json:"md5,omitempty"`
}

// ListFilesResponse represents the response from the list files API
type ListFilesResponse struct {
	Errno     int        `json:"errno"`
	GuidInfo  string     `json:"guid_info"`
	List      []FileInfo `json:"list"`
	RequestID int64      `json:"request_id"`
	Guid      int        `json:"guid"`
}

// PrecreateResponse represents the response from the precreate API
type PrecreateResponse struct {
	Errno      int    `json:"errno"`
	UploadID   string `json:"uploadid"`
	BlockList  []int  `json:"block_list"`
	RequestID  int64  `json:"request_id"`
	ReturnType int    `json:"return_type"` // 1: need upload, 2: no need upload (file already exists and matches)
}

// CreateFileResponse represents the response from the create file API
type CreateFileResponse struct {
	Errno          int    `json:"errno"`
	FsID           int64  `json:"fs_id"`
	Path           string `json:"path"`
	CTime          int64  `json:"ctime"`
	MTime          int64  `json:"mtime"`
	MD5            string `json:"md5"`
	Size           int64  `json:"size"`
	Isdir          int    `json:"isdir"`
	Ifhassubdir    int    `json:"ifhassubdir"`
	Category       int    `json:"category"`
	ServerFilename string `json:"server_filename"`
	ParentPath     string `json:"parent_path"`
}


// TokenFile represents the structure for storing tokens in a file
type TokenFile struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	UID          string    `json:"uid"`
	CreatedAt    time.Time `json:"created_at"` // Time when token was obtained
}

// Client represents a Baidu Pan client
type Client struct {
	client         *http.Client
	downloadClient *http.Client // Separate client with longer timeout for downloads
	clientID       string
	clientSecret   string
	accessToken    string
	refreshToken   string
	expiresIn      int
	uid            string
	tokenFile      string
	tokenCreatedAt time.Time // Time when the current tokens were obtained
}

// NewClient creates a new Baidu Pan client
func NewClient(clientID, clientSecret, tokenPath string) *Client {
	// Ensure the directory exists
	tokenDir := filepath.Dir(tokenPath)
	os.MkdirAll(tokenDir, 0755)

	return &Client{
		client:         &http.Client{Timeout: 30 * time.Second},
		downloadClient: &http.Client{Timeout: 300 * time.Second}, // 5 minutes timeout for downloads
		clientID:       clientID,
		clientSecret:   clientSecret,
		tokenFile:      tokenPath,
	}
}

// GetDeviceCode initiates the device code flow
func (c *Client) GetDeviceCode() (*DeviceCodeResponse, error) {
	params := url.Values{}
	params.Add("client_id", c.clientID)
	params.Add("response_type", "device_code")
	params.Add("scope", "basic,netdisk")

	req, err := http.NewRequest("POST", deviceCodeURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deviceCodeResp DeviceCodeResponse
	err = json.Unmarshal(body, &deviceCodeResp)
	if err != nil {
		return nil, err
	}

	return &deviceCodeResp, nil
}

// PollForToken polls the token endpoint until the user completes authorization
func (c *Client) PollForToken(deviceCode string) (*TokenResponse, error) {
	params := url.Values{}
	params.Add("grant_type", "device_token")
	params.Add("code", deviceCode)
	params.Add("client_id", c.clientID)
	params.Add("client_secret", c.clientSecret)

	// Get initial device code response to know poll interval and timeout
	initialResp, err := c.GetDeviceCodeForPoll(deviceCode)
	if err != nil {
		return nil, err
	}

	interval := initialResp.Interval
	if interval == 0 {
		interval = 5 // default to 5 seconds if not provided
	}

	// Calculate timeout (expires_in seconds)
	timeout := time.Now().Add(time.Duration(initialResp.ExpiresIn) * time.Second)

	for {
		// Check if timeout has been reached
		if time.Now().After(timeout) {
			return nil, fmt.Errorf("device code has expired")
		}

		tokenResp, err := c.requestToken(params)
		if err != nil {
			return nil, err
		}

		// Check if we got an access token (authorization completed)
		if tokenResp.AccessToken != "" {
			c.accessToken = tokenResp.AccessToken
			c.refreshToken = tokenResp.RefreshToken
			c.expiresIn = tokenResp.ExpiresIn
			c.uid = tokenResp.UID
			return tokenResp, nil
		}

		// Wait for next poll interval
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// GetDeviceCodeForPoll gets device code info specifically for polling (internal use)
func (c *Client) GetDeviceCodeForPoll(deviceCode string) (*DeviceCodeResponse, error) {
	// This would be a separate call to get polling parameters
	// For now we'll return a response with the default values
	return &DeviceCodeResponse{
		DeviceCode: deviceCode,
		Interval:   5,    // Default interval
		ExpiresIn:  1800, // Default 30 minutes
	}, nil
}

// requestToken makes the token request
func (c *Client) requestToken(params url.Values) (*TokenResponse, error) {
	req, err := http.NewRequest("POST", accessTokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		// Check if it's an authorization pending response
		if resp.StatusCode == http.StatusBadRequest {
			var errorResp struct {
				Error            string `json:"error"`
				ErrorDescription string `json:"error_description"`
			}
			json.Unmarshal(body, &errorResp)

			if errorResp.Error == "authorization_pending" {
				// This is normal during polling, return empty token to continue polling
				return &TokenResponse{}, nil
			} else if errorResp.Error == "slow_down" {
				// The polling interval is too fast, return empty token to continue with increased interval
				return &TokenResponse{}, nil
			}
		}

		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// Authorize tries to load existing tokens first, and if not available or valid, performs device code authorization
func (c *Client) Authorize(ctx context.Context) error {
	// Try to load existing tokens first
	if c.HasValidToken() {
		err := c.LoadTokens()
		if err == nil {
			// Check if token is expired or will expire soon (within 2 days)
			if c.IsTokenExpired() {
				PrintSuccess("Access token is expired or will expire soon, attempting to refresh...")

				// Try to refresh the token
				refreshErr := c.RefreshToken()
				if refreshErr != nil {
					PrintError(fmt.Sprintf("Token refresh failed: %v", refreshErr))
					PrintSuccess("Removing expired token file and starting new authorization...")

					// If refresh fails, remove the token file and start new authorization
					os.Remove(c.tokenFile)

					// Now perform device code authorization
					return c.performDeviceCodeAuth(ctx)
				} else {
					PrintSuccess("Token refreshed successfully!")
					return nil
				}
			} else {
				PrintSuccess("Using existing tokens from .bdfs_certs")
				return nil
			}
		} else {
			PrintError(fmt.Sprintf("Could not load existing tokens, will re-authorize: %v", err))
		}
	}

	// If no existing tokens or loading failed, perform device code authorization
	return c.performDeviceCodeAuth(ctx)
}

// performDeviceCodeAuth performs the actual device code authorization flow
func (c *Client) performDeviceCodeAuth(ctx context.Context) error {
	deviceResp, err := c.GetDeviceCode()
	if err != nil {
		return fmt.Errorf("failed to get device code: %w", err)
	}

	fmt.Printf("Please visit: %s\n", deviceResp.VerificationURL)
	fmt.Printf("Enter the code: %s\n", deviceResp.UserCode)
	fmt.Printf("The code will expire in %d seconds.\n", deviceResp.ExpiresIn)

	// Start polling for token in a goroutine with context cancellation
	tokenChan := make(chan *TokenResponse, 1)
	errChan := make(chan error, 1)

	go func() {
		token, err := c.PollForToken(deviceResp.DeviceCode)
		if err != nil {
			errChan <- err
			return
		}
		tokenChan <- token
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case token := <-tokenChan:
		c.accessToken = token.AccessToken
		c.refreshToken = token.RefreshToken
		c.expiresIn = token.ExpiresIn
		c.uid = token.UID

		// Save tokens to file
		err = c.SaveTokens()
		if err != nil {
			return fmt.Errorf("failed to save tokens: %w", err)
		}

		PrintSuccess("Authorization successful! Tokens saved to .bdfs_certs")
		return nil
	case err := <-errChan:
		return fmt.Errorf("failed to get token: %w", err)
	}
}

// SaveTokens saves the access token to a file
func (c *Client) SaveTokens() error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token to save")
	}

	tokenFile := &TokenFile{
		AccessToken:  c.accessToken,
		RefreshToken: c.refreshToken,
		ExpiresIn:    c.expiresIn,
		UID:          c.uid,
		CreatedAt:    time.Now(), // Set current time when saving
	}

	// Update the client's token creation time as well
	c.tokenCreatedAt = tokenFile.CreatedAt

	data, err := json.MarshalIndent(tokenFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.tokenFile, data, 0600)
}

// LoadTokens loads the access token from a file
func (c *Client) LoadTokens() error {
	data, err := os.ReadFile(c.tokenFile)
	if err != nil {
		return fmt.Errorf("unable to read token file: %w", err)
	}

	var tokenFile TokenFile
	err = json.Unmarshal(data, &tokenFile)
	if err != nil {
		return err
	}

	c.accessToken = tokenFile.AccessToken
	c.refreshToken = tokenFile.RefreshToken
	c.expiresIn = tokenFile.ExpiresIn
	c.uid = tokenFile.UID
	c.tokenCreatedAt = tokenFile.CreatedAt

	return nil
}

// HasValidToken checks if there's a valid token in the file
func (c *Client) HasValidToken() bool {
	_, err := os.Stat(c.tokenFile)
	return err == nil
}


// IsTokenExpired checks if the token is expired or will expire soon (within 2 days)
func (c *Client) IsTokenExpired() bool {
	if c.tokenCreatedAt.IsZero() || c.expiresIn == 0 {
		// If we don't have creation time or expiration info, assume it's expired
		return true
	}

	// Calculate when the token will expire (expires_in is in seconds)
	expirationTime := c.tokenCreatedAt.Add(time.Duration(c.expiresIn) * time.Second)
	twoDaysBefore := time.Now().Add(48 * time.Hour) // 2 days before expiration

	// Return true if token expires within 2 days or has already expired
	return expirationTime.Before(time.Now()) || expirationTime.Before(twoDaysBefore)
}

// RefreshToken attempts to refresh the access token using the refresh token
func (c *Client) RefreshToken() error {
	if c.refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", c.refreshToken)
	params.Add("client_id", c.clientID)
	params.Add("client_secret", c.clientSecret)

	req, err := http.NewRequest("POST", accessTokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return err
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("refresh returned empty access token")
	}

	// Update client with new tokens
	c.accessToken = tokenResp.AccessToken
	c.refreshToken = tokenResp.RefreshToken
	c.expiresIn = tokenResp.ExpiresIn
	c.uid = tokenResp.UID

	// Save the refreshed tokens
	return c.SaveTokens()
}

// CalculateMD5 calculates the MD5 hash of a given file
func CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5 for file %s: %w", filePath, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CalculateSliceMD5 calculates MD5 hashes for fixed-size slices of a file
func CalculateSliceMD5(filePath string, sliceSize int64) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var md5List []string
	buffer := make([]byte, sliceSize)
	for {
		n, err := file.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read file slice: %w", err)
		}

		hash := md5.New()
		hash.Write(buffer[:n])
		md5List = append(md5List, fmt.Sprintf("%x", hash.Sum(nil)))

		if err == io.EOF {
			break
		}
	}

	return md5List, nil
}

// EnsureRemoteDirExists verifies the remote directory path is valid.
// Baidu Pan's API typically creates parent directories if they don't exist during upload,
// so this function primarily serves to validate the path format.
func (c *Client) EnsureRemoteDirExists(remotePath string) error {
	if !strings.HasPrefix(remotePath, "/") {
		return fmt.Errorf("remote path must be an absolute path starting with '/'")
	}
	return nil
}

// GetSourceFileName extracts the filename from a path
func GetSourceFileName(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// HasRefreshToken returns whether the client has a refresh token available
func (c *Client) HasRefreshToken() bool {
	return c.refreshToken != ""
}
