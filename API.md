# Baidu Pan API Documentation

This document provides detailed information about the Baidu Pan API client implemented in the `pan` package.

## Table of Contents

1. [Client Structure](#client-structure)
2. [Authentication](#authentication)
3. [File Operations](#file-operations)
4. [Directory Operations](#directory-operations)
5. [Utility Functions](#utility-functions)
6. [Response Types](#response-types)

## Client Structure

The `Client` struct represents a Baidu Pan client and provides methods for interacting with the Baidu Pan API.

```go
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
```

## Authentication

### NewClient
```go
func NewClient(clientID, clientSecret, tokenPath string) *Client
```
Creates a new Baidu Pan client with the provided client ID, client secret, and token file path. Sets up HTTP clients with appropriate timeouts.

### GetDeviceCode
```go
func (c *Client) GetDeviceCode() (*DeviceCodeResponse, error)
```
Initiates the device code flow for authentication. Returns a `DeviceCodeResponse` containing device code, user code, and verification URL.

### PollForToken
```go
func (c *Client) PollForToken(deviceCode string) (*TokenResponse, error)
```
Polls the token endpoint until the user completes authorization. Continues polling until either a token is received or the device code expires.

### Authorize
```go
func (c *Client) Authorize(ctx context.Context) error
```
Tries to load existing tokens first, and if not available or valid, performs device code authorization. If tokens exist but are expired, attempts to refresh them.

### HasValidToken
```go
func (c *Client) HasValidToken() bool
```
Checks if there's a valid token in the file.

### LoadTokens
```go
func (c *Client) LoadTokens() error
```
Loads the access token from a file.

### SaveTokens
```go
func (c *Client) SaveTokens() error
```
Saves the access token to a file.

### IsTokenExpired
```go
func (c *Client) IsTokenExpired() bool
```
Checks if the token is expired or will expire soon (within 2 days).

### RefreshToken
```go
func (c *Client) RefreshToken() error
```
Attempts to refresh the access token using the refresh token.


### HasRefreshToken
```go
func (c *Client) HasRefreshToken() bool
```
Returns whether the client has a refresh token available.

## File Operations

### ListFiles
```go
func (c *Client) ListFiles(dirPath string) ([]FileInfo, error)
```
Lists files in a specified directory. Returns a slice of `FileInfo` structs representing the files and directories in the specified path.

### GetFileInfo
```go
func (c *Client) GetFileInfo(filePath string) (*FileInfo, error)
```
Gets information about a specific file by listing files in the parent directory and finding the target file.

### GetFileInfoByPath
```go
func (c *Client) GetFileInfoByPath(filePath string) (*FileInfo, error)
```
Gets information about a specific file or directory by its path using the list API with a filename filter.

### GetDetailedFileInfo
```go
func (c *Client) GetDetailedFileInfo(filePath string) (*FileInfo, error)
```
Gets detailed information about a file using the meta API, which is more efficient than listing files when only info about one file is needed.

### DownloadFile
```go
func (c *Client) DownloadFile(filePath string) (*http.Response, error)
```
Downloads a file from Baidu Pan, returning an HTTP response. Uses a client with a longer timeout for downloads.

### DownloadFileToPath
```go
func (c *Client) DownloadFileToPath(filePath, localPath string) error
```
Downloads a file from Baidu Pan and saves it to the specified local path. Includes progress reporting.

### ReadFileContent
```go
func (c *Client) ReadFileContent(filePath string) ([]byte, error)
```
Reads the content of a file from Baidu Pan directly into memory.

### UploadFile
```go
func (c *Client) UploadFile(localFilePath, remoteFilePath string) error
```
Uploads a local file to Baidu Pan using the multi-step upload process:
1. Calculate slice MD5s
2. Call precreate API
3. Upload file slices
4. Call create file API to finalize

### RemoveFile
```go
func (c *Client) RemoveFile(filePath string) error
```
Removes a single file or directory from Baidu Pan.

### RemoveFiles
```go
func (c *Client) RemoveFiles(filePaths []string) error
```
Removes multiple files or directories from Baidu Pan in a single operation.

### MoveFile
```go
func (c *Client) MoveFile(sourcePath, destDir string) error
```
Moves a single file or directory from source path to destination directory in Baidu Pan.

### MoveFiles
```go
func (c *Client) MoveFiles(moveRequests []MoveRequest) error
```
Moves multiple files based on the provided `MoveRequest` structs.

### CopyFile
```go
func (c *Client) CopyFile(sourcePath, destPath string) error
```
Copies a single file or directory from source path to destination directory in Baidu Pan.

### CopyFiles
```go
func (c *Client) CopyFiles(copyRequests []CopyRequest) error
```
Copies multiple files based on the provided `CopyRequest` structs.

### RenameFile
```go
func (c *Client) RenameFile(sourcePath, newName string) error
```
Renames a single file or directory in Baidu Pan.

### RenameFiles
```go
func (c *Client) RenameFiles(renameRequests []RenameRequest) error
```
Renames multiple files based on the provided `RenameRequest` structs.

## Directory Operations

### CreateDir
```go
func (c *Client) CreateDir(remotePath string) error
```
Creates a directory in Baidu Pan at the specified remote path.

### Walk
```go
func (c *Client) Walk(rootPath string) (<-chan FileInfo, <-chan error)
```
Recursively walks through directories and files starting from the specified root path. Returns two channels: one for file information and one for errors.

### WalkRecursive
```go
func (c *Client) WalkRecursive(path string, fileChan chan<- FileInfo, errChan chan<- error)
```
Internal recursive method for walking through directories and files, called by `Walk`.

## Utility Functions

### CalculateMD5
```go
func CalculateMD5(filePath string) (string, error)
```
Calculates the MD5 hash of a given file.

### CalculateSliceMD5
```go
func CalculateSliceMD5(filePath string, sliceSize int64) ([]string, error)
```
Calculates MD5 hashes for fixed-size slices of a file.

### EnsureRemoteDirExists
```go
func (c *Client) EnsureRemoteDirExists(remotePath string) error
```
Verifies the remote directory path is valid. Baidu Pan's API typically creates parent directories if they don't exist during upload, so this function primarily serves to validate the path format.

### GetSourceFileName
```go
func GetSourceFileName(path string) string
```
Extracts the filename from a path.

### GetDiskInfo
```go
func (c *Client) GetDiskInfo() (*DiskInfoResponse, error)
```
Gets the user's cloud storage usage information, including total space, used space, free space, and expiration status.

### FormatDiskInfo
```go
func FormatDiskInfo(info *DiskInfoResponse) string
```
Formats the disk information in a human-readable way.

### FormatBytes
```go
func FormatBytes(bytes int64) string
```
Converts bytes to a human-readable format (e.g., KB, MB, GB).

### FormatFileInfo
```go
func FormatFileInfo(fileInfo *FileInfo) string
```
Formats the file information in a human-readable way.

### MapFileType
```go
func MapFileType(isDir int) string
```
Converts the `isdir` field to a readable file type (either "Directory" or "File").

### FormatTime
```go
func FormatTime(unixTime int64) string
```
Converts Unix timestamp to readable time format.

### PrintSuccess
```go
func PrintSuccess(message string)
```
Prints a success message with consistent formatting.

### PrintError
```go
func PrintError(message string)
```
Prints an error message with consistent formatting.

### PrintErrorAndExit
```go
func PrintErrorAndExit(message string)
```
Prints an error message and exits with code 1.

### GetErrorMessage
```go
func GetErrorMessage(errno int) string
```
Returns a human-readable error message for common errno values.

### GetRenameErrorMessage
```go
func GetRenameErrorMessage(errno int) string
```
Returns a human-readable error message for common errno values in rename operations.

### GetMoveErrorMessage
```go
func GetMoveErrorMessage(errno int) string
```
Returns a human-readable error message for common errno values in move operations.

### GetCopyErrorMessage
```go
func GetCopyErrorMessage(errno int) string
```
Returns a human-readable error message for common errno values in copy operations.

## Response Types

### DeviceCodeResponse
Represents the response from device code endpoint.
```go
type DeviceCodeResponse struct {
    DeviceCode      string `json:"device_code"`
    UserCode        string `json:"user_code"`
    VerificationURL string `json:"verification_url"`
    ExpiresIn       int    `json:"expires_in"`
    Interval        int    `json:"interval"`
    QRCode          string `json:"qrcode,omitempty"`
}
```

### TokenResponse
Represents the access token response.
```go
type TokenResponse struct {
    AccessToken   string `json:"access_token"`
    RefreshToken  string `json:"refresh_token"`
    ExpiresIn     int    `json:"expires_in"`
    Scope         string `json:"scope"`
    SessionKey    string `json:"session_key"`
    SessionSecret string `json:"session_secret"`
    UID           string `json:"uid"`
}
```

### FileInfo
Represents a file or directory in Baidu Pan.
```go
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
```

### ListFilesResponse
Represents the response from the list files API.
```go
type ListFilesResponse struct {
    Errno     int        `json:"errno"`
    GuidInfo  string     `json:"guid_info"`
    List      []FileInfo `json:"list"`
    RequestID int64      `json:"request_id"`
    Guid      int        `json:"guid"`
}
```

### PrecreateResponse
Represents the response from the precreate API.
```go
type PrecreateResponse struct {
    Errno      int    `json:"errno"`
    UploadID   string `json:"uploadid"`
    BlockList  []int  `json:"block_list"`
    RequestID  int64  `json:"request_id"`
    ReturnType int    `json:"return_type"` // 1: need upload, 2: no need upload (file already exists and matches)
}
```

### CreateFileResponse
Represents the response from the create file API.
```go
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
```

### DeleteResponse
Represents the response from the delete API.
```go
type DeleteResponse struct {
    Errno     int           `json:"errno"`
    RequestID int64         `json:"request_id"`
    List      []DeleteEntry `json:"list"`
}
```

### DeleteEntry
Represents the result for each file in the delete operation.
```go
type DeleteEntry struct {
    Path  string `json:"path"`
    Errno int    `json:"errno"`
}
```

### RenameResponse
Represents the response from the rename API.
```go
type RenameResponse struct {
    Errno     int          `json:"errno"`
    Info      []RenameInfo `json:"info"`
    TaskID    int64        `json:"taskid"`
    RequestID int64        `json:"request_id"`
}
```

### RenameInfo
Represents the result for each renamed file in the response.
```go
type RenameInfo struct {
    Path  string `json:"path"`
    Errno int    `json:"errno"`
}
```

### RenameRequest
Represents the structure for a file to be renamed.
```go
type RenameRequest struct {
    Path    string `json:"path"`
    NewName string `json:"newname"`
}
```

### MoveResponse
Represents the response from the move API.
```go
type MoveResponse struct {
    Errno     int        `json:"errno"`
    Info      []MoveInfo `json:"info"`
    TaskID    int64      `json:"taskid"`
    RequestID int64      `json:"request_id"`
}
```

### MoveInfo
Represents the result for each moved file in the response.
```go
type MoveInfo struct {
    Path    string `json:"path"`
    Dest    string `json:"dest"`
    NewName string `json:"newname"`
    Errno   int    `json:"errno"`
}
```

### MoveRequest
Represents the structure for a file to be moved.
```go
type MoveRequest struct {
    Path    string `json:"path"`
    Dest    string `json:"dest"`
    NewName string `json:"newname"`
}
```

### CopyResponse
Represents the response from the copy API.
```go
type CopyResponse struct {
    Errno     int        `json:"errno"`
    Info      []CopyInfo `json:"info"`
    TaskID    int64      `json:"taskid"`
    RequestID int64      `json:"request_id"`
}
```

### CopyInfo
Represents the result for each copied file in the response.
```go
type CopyInfo struct {
    Path    string `json:"path"`
    Dest    string `json:"dest"`
    NewName string `json:"newname"`
    Errno   int    `json:"errno"`
}
```

### CopyRequest
Represents the structure for a file to be copied.
```go
type CopyRequest struct {
    Path    string `json:"path"`
    Dest    string `json:"dest"`
    NewName string `json:"newname"`
}
```

### DiskInfoResponse
Represents the response from the disk info API.
```go
type DiskInfoResponse struct {
    Errno     int64  `json:"errno"`
    Total     int64  `json:"total"`     // Total space size in bytes
    Used      int64  `json:"used"`      // Used space size in bytes
    Free      int64  `json:"free"`      // Free capacity in bytes
    Expire    bool   `json:"expire"`    // Whether capacity will expire within 7 days
    RequestID int64  `json:"request_id"`
}
```

### Config
Structure for configuration loading from environment variables or TOML file.
```go
type Config struct {
    ClientID     string `toml:"client_id"`
    ClientSecret string `toml:"client_secret"`
    TokenPath    string `toml:"token_path"`
}
```

### TokenFile
Structure for storing tokens in a file.
```go
type TokenFile struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresIn    int       `json:"expires_in"`
    UID          string    `json:"uid"`
    CreatedAt    time.Time `json:"created_at"` // Time when token was obtained
}
```

### ProgressWriter
A wrapper around an io.Writer that reports download progress.
```go
type ProgressWriter struct {
    writer     io.Writer
    totalSize  int64
    downloaded int64
    fileName   string
}
```