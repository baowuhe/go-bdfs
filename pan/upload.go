package pan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// UploadFile uploads a local file to Baidu Pan
func (c *Client) UploadFile(localFilePath, remoteFilePath string) error {
	if c.accessToken == "" {
		return fmt.Errorf("no access token, please authorize first")
	}

	// 1. Get local file information
	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to get local file info: %w", err)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("cannot upload directory, please specify a file: %s", localFilePath)
	}

	fileSize := fileInfo.Size()
	fileName := fileInfo.Name()

	// Ensure remote path is valid
	if err := c.ensureRemoteDirExists(filepath.Dir(remoteFilePath)); err != nil {
		return err
	}

	// Calculate slice MD5s (Baidu typically uses 4MB slices)
	const sliceSize = 4 * 1024 * 1024 // 4MB
	sliceMD5s, err := calculateSliceMD5(localFilePath, sliceSize)
	if err != nil {
		return fmt.Errorf("failed to calculate slice MD5s: %w", err)
	}

	// Convert slice MD5s to JSON string for precreate API
	sliceMD5sJSON, err := json.Marshal(sliceMD5s)
	if err != nil {
		return fmt.Errorf("failed to marshal slice MD5s to JSON: %w", err)
	}

	PrintSuccess(fmt.Sprintf("Uploading %s (%s) to %s", fileName, byteCountToHumanReadable(fileSize), remoteFilePath))

	// 2. Call Precreate API
	precreateParams := url.Values{}
	precreateParams.Add("access_token", c.accessToken)
	precreateParams.Add("path", remoteFilePath)
	precreateParams.Add("size", fmt.Sprintf("%d", fileSize))
	precreateParams.Add("isdir", "0")    // 0 for file
	precreateParams.Add("autoinit", "1") // Let Baidu initiate the upload
	precreateParams.Add("rtype", "1")    // Overwrite existing file
	precreateParams.Add("block_list", string(sliceMD5sJSON))

	precreateReq, err := http.NewRequest("POST", uploadPrecreateURL, strings.NewReader(precreateParams.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create precreate request: %w", err)
	}
	precreateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	precreateResp, err := c.client.Do(precreateReq)
	if err != nil {
		return fmt.Errorf("precreate request failed: %w", err)
	}
	defer precreateResp.Body.Close()

	precreateBody, err := io.ReadAll(precreateResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read precreate response body: %w", err)
	}

	if precreateResp.StatusCode != http.StatusOK {
		return fmt.Errorf("precreate API failed with status %d: %s", precreateResp.StatusCode, string(precreateBody))
	}

	var precreateResponse PrecreateResponse
	err = json.Unmarshal(precreateBody, &precreateResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal precreate response: %w", err)
	}

	if precreateResponse.Errno != 0 {
		return fmt.Errorf("precreate API returned error code %d: %s", precreateResponse.Errno, string(precreateBody))
	}

	// 3. Handle Precreate Response
	if precreateResponse.ReturnType == 2 {
		PrintSuccess(fmt.Sprintf("File '%s' already exists on Baidu Pan and matches the local file. Skipping upload.", remoteFilePath))
		return nil
	}

	if precreateResponse.UploadID == "" {
		return fmt.Errorf("precreate API did not return uploadid")
	}

	// 4. Upload Slices
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file for uploading: %w", err)
	}
	defer localFile.Close()

	PrintSuccess("Starting slice upload...")
	uploadedBytes := int64(0)

	// Create a buffer for reading file slices
	sliceBuffer := make([]byte, sliceSize)

	for i := 0; i < len(sliceMD5s); i++ {
		// Calculate the starting offset for the current slice
		offset := int64(i) * sliceSize
		_, err := localFile.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek to slice position: %w", err)
		}

		// Read the current slice into the buffer
		n, err := io.ReadFull(localFile, sliceBuffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("failed to read file slice %d: %w", i, err)
		}
		currentSlice := sliceBuffer[:n]

		// Create multipart form data for slice upload
		var requestBody bytes.Buffer
		multipartWriter := multipart.NewWriter(&requestBody)

		// Add "file" field
		fileWriter, err := multipartWriter.CreateFormFile("file", fileName)
		if err != nil {
			return fmt.Errorf("failed to create form file for slice: %w", err)
		}
		_, err = fileWriter.Write(currentSlice)
		if err != nil {
			return fmt.Errorf("failed to write slice data to form file: %w", err)
		}

		// Close the multipart writer to finalize the form data
		multipartWriter.Close()

		sliceUploadURL := fmt.Sprintf("%s?access_token=%s&method=upload&type=tmpfile&path=%s&uploadid=%s&partseq=%d",
			uploadSuperfileURL, c.accessToken, remoteFilePath, precreateResponse.UploadID, i)

		sliceUploadReq, err := http.NewRequest("POST", sliceUploadURL, &requestBody)
		if err != nil {
			return fmt.Errorf("failed to create slice upload request: %w", err)
		}
		sliceUploadReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())

		sliceUploadResp, err := c.client.Do(sliceUploadReq)
		if err != nil {
			return fmt.Errorf("slice upload request failed for part %d: %w", i, err)
		}
		defer sliceUploadResp.Body.Close()

		sliceUploadBody, err := io.ReadAll(sliceUploadResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read slice upload response body for part %d: %w", i, err)
		}

		if sliceUploadResp.StatusCode != http.StatusOK {
			return fmt.Errorf("slice upload API failed for part %d with status %d: %s", i, sliceUploadResp.StatusCode, string(sliceUploadBody))
		}

		uploadedBytes += int64(n)
		fmt.Printf("\r%d / %d (%.2f%%)",
			uploadedBytes,
			fileSize,
			float64(uploadedBytes)/float64(fileSize)*100)
		os.Stdout.Sync()
	}
	PrintSuccess("All slices uploaded.")

	// 5. Call Create File API to finalize
	createFileParams := url.Values{}
	createFileParams.Add("access_token", c.accessToken)
	createFileParams.Add("path", remoteFilePath)
	createFileParams.Add("size", fmt.Sprintf("%d", fileSize))
	createFileParams.Add("isdir", "0")
	createFileParams.Add("uploadid", precreateResponse.UploadID)
	createFileParams.Add("block_list", string(sliceMD5sJSON)) // Need to send all block MD5s again
	createFileParams.Add("rtype", "1")                        // Overwrite existing file

	createFileReq, err := http.NewRequest("POST", uploadCreateFileUrl, strings.NewReader(createFileParams.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create create file request: %w", err)
	}
	createFileReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	createFileResp, err := c.client.Do(createFileReq)
	if err != nil {
		return fmt.Errorf("create file request failed: %w", err)
	}
	defer createFileResp.Body.Close()

	createFileBody, err := io.ReadAll(createFileResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read create file response body: %w", err)
	}

	if createFileResp.StatusCode != http.StatusOK {
		return fmt.Errorf("create file API failed with status %d: %s", createFileResp.StatusCode, string(createFileBody))
	}

	var createFileResponse CreateFileResponse
	err = json.Unmarshal(createFileBody, &createFileResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal create file response: %w", err)
	}

	if createFileResponse.Errno != 0 {
		return fmt.Errorf("create file API returned error code %d: %s", createFileResponse.Errno, string(createFileBody))
	}

	PrintSuccess(fmt.Sprintf("File '%s' uploaded successfully to Baidu Pan as '%s'", fileName, createFileResponse.Path))

	return nil
}

// byteCountToHumanReadable converts bytes to a human-readable string
func byteCountToHumanReadable(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
