package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	pan "github.com/baowuhe/go-bdfs/pan"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"
)

const VERSION = "v0.1.2"

// Config represents the configuration structure
type Config struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	TokenPath    string `toml:"token_path"`
}

// LoadConfig loads configuration from environment variables or TOML file
func LoadConfig() (*Config, error) {
	config := &Config{}

	// First, try to get config file path from environment variable
	configFilePath := os.Getenv("BDFS_CONFIG_FILE_PATH")
	if configFilePath == "" {
		// If BDFS_CONFIG_FILE_PATH is not set, use default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configFilePath = filepath.Join(homeDir, ".local", "app", "bdfs", "config.toml")
	}

	// Try to load from config file first
	if _, err := os.Stat(configFilePath); err == nil {
		// Config file exists, read it
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		err = toml.Unmarshal(data, config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	} else {
		// Config file doesn't exist, try loading from environment variables
		config.ClientID = os.Getenv("BDFS_CLIENT_ID")
		config.ClientSecret = os.Getenv("BDFS_CLIENT_SECRET")
		config.TokenPath = os.Getenv("BDFS_TOKEN_PATH")
	}

	// Validate that all required parameters are provided
	if config.ClientID == "" || config.ClientSecret == "" || config.TokenPath == "" {
		return nil, fmt.Errorf("missing required configuration parameters. Please set either:\n" +
			"  1. BDFS_CONFIG_FILE_PATH environment variable pointing to a TOML file with client_id, client_secret, and token_path, or\n" +
			"  2. BDFS_CLIENT_ID, BDFS_CLIENT_SECRET, and BDFS_TOKEN_PATH environment variables")
	}

	return config, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("go-bdfs: Baidu Pan client")
		fmt.Println("Usage: go-bdfs <command> [arguments]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  ls          List files in a directory")
		fmt.Println("  dl          Download a file from Baidu Pan")
		fmt.Println("  ul          Upload a file to Baidu Pan")
		fmt.Println("  rm          Remove a file or directory from Baidu Pan")
		fmt.Println("  mv          Move a file or directory to another directory in Baidu Pan")
		fmt.Println("  rn          Rename a file or directory in Baidu Pan")
		fmt.Println("  md          Create a directory in Baidu Pan")
		fmt.Println("  cp          Copy a file or directory in Baidu Pan")
		fmt.Println("  if          Get information about a file in Baidu Pan")
		fmt.Println("  di          Get disk information (storage usage) from Baidu Pan")
		fmt.Println("  ar          Refresh the access token using the refresh token")
		fmt.Println("  version     Show the version information")
		fmt.Println("")
		fmt.Println("Use 'go-bdfs <command> -h' for more information about a command.")
		os.Exit(1)
	}

	// Parse command
	cmd := os.Args[1]

	// Handle commands that don't require authorization first
	switch strings.ToLower(cmd) {
	case "version":
		versionCommand()
		return
	case "help", "-h", "--help":
		showHelp()
		return
	}

	// Load configuration from environment variables or TOML file
	config, err := LoadConfig()
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error loading configuration: %v", err))
		fmt.Println("You can set BDFS_CLIENT_ID, BDFS_CLIENT_SECRET, and BDFS_TOKEN_PATH environment variables")
		fmt.Println("Or create a config file at $HOME/.local/app/bdfs/config.toml with the following format:")
		fmt.Println("")
		fmt.Println("Format (direct values):")
		fmt.Println("client_id = \"your_client_id\"")
		fmt.Println("client_secret = \"your_client_secret\"")
		fmt.Println("token_path = \"path/to/your/token/file\"")
		fmt.Println("")
		fmt.Println("Alternatively, set the BDFS_CONFIG_FILE_PATH environment variable to point to your config file")
		fmt.Println("[×] Launch failed!ss")
		os.Exit(1)
	}

	// For all other commands, load the client and perform authorization
	client := pan.NewClient(config.ClientID, config.ClientSecret, config.TokenPath)

	// Set a timeout for authorization
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Starting Baidu Pan authorization...")

	// Try to load existing tokens or perform device code authorization
	err = client.Authorize(ctx)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Authorization failed: %v", err))
		os.Exit(1)
	}

	// Execute requested command
	switch strings.ToLower(cmd) {
	case "ls":
		listCommand(client)
	case "dl":
		downloadCommand(client)
	case "ul":
		uploadCommand(client)
	case "rm":
		removeCommand(client)
	case "mv":
		moveCommand(client)
	case "rn":
		renameCommand(client)
	case "md":
		mkdirCommand(client)
	case "cp":
		copyCommand(client)
	case "if":
		infoCommand(client)
	case "di":
		diskInfoCommand(client)
	case "ar":
		refreshTokenCommand(client)
	default:
		pan.PrintError(fmt.Sprintf("Unknown command: %s", cmd))
		fmt.Println("Run 'go-bdfs' for usage information.")
		os.Exit(1)
	}
}

func listCommand(client *pan.Client) {
	// Create a new flag set for the list command using pflag
	listFlags := pflag.NewFlagSet("ls", pflag.ExitOnError)
	var dir string
	var help bool

	listFlags.StringVarP(&dir, "path", "p", "/", "Directory to list (default: /)")
	listFlags.BoolVarP(&help, "help", "h", false, "Show help for list command")

	// Parse flags starting from os.Args[2] (after the 'list' command)
	if err := listFlags.Parse(os.Args[2:]); err != nil {
		// Error already handled by pflag.ExitOnError
		return
	}

	// Show help if -h or --help flag is provided
	if help {
		listFlags.PrintDefaults()
		return
	}

	pan.PrintSuccess(fmt.Sprintf("Listing files in directory: %s", dir))

	files, err := client.ListFiles(dir)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error listing files: %v", err))
		os.Exit(1)
	}

	if len(files) == 0 {
		pan.PrintSuccess("No files found.")
		return
	}

	// Sort files by filename in ascending order
	sort.Slice(files, func(i, j int) bool {
		return files[i].ServerFilename < files[j].ServerFilename
	})

	// Print files with the new format: <类型> | <文件名> | <文件路径> | <文件大小> | <创建时间> | <更新时间>
	for _, file := range files {
		// Determine file type: D for directory, F for file
		fileType := "F"
		if file.IsDir == 1 {
			fileType = "D"
		}

		// Format file size - use "-" for directories
		sizeStr := fmt.Sprintf("%d", file.Size)
		if file.IsDir == 1 {
			sizeStr = "-"
		}

		// Format creation and update times from Unix timestamps
		ctime := time.Unix(file.ServerCtime, 0)
		mtime := time.Unix(file.ServerMtime, 0)

		// Output in the required format
		fmt.Printf("%s | %s | %s | %s | %s | %s\n",
			fileType,
			file.ServerFilename,
			file.Path,
			sizeStr,
			ctime.Format("2006-01-02 15:04:05"),
			mtime.Format("2006-01-02 15:04:05"))
	}
}

func downloadCommand(client *pan.Client) {
	// Create a new flag set for the download command using pflag
	downloadFlags := pflag.NewFlagSet("dl", pflag.ExitOnError)
	var filePath string
	var outputPath string
	var help bool

	downloadFlags.StringVarP(&filePath, "source", "s", "", "File path in Baidu Pan to download (required)")
	downloadFlags.StringVarP(&outputPath, "destination", "d", "", "Local output file path (optional, defaults to current directory with original filename)")
	downloadFlags.BoolVarP(&help, "help", "h", false, "Show help for download command")

	// Parse flags starting from os.Args[2] (after the 'download' command)
	if err := downloadFlags.Parse(os.Args[2:]); err != nil {
		// Error already handled by pflag.ExitOnError
		return
	}

	// Show help if -h or --help flag is provided
	if help {
		downloadFlags.PrintDefaults()
		return
	}

	// Check if file path is provided
	if filePath == "" {
		pan.PrintError("Error: -f or --file flag is required to specify the file to download")
		downloadFlags.PrintDefaults()
		os.Exit(1)
	}

	// Determine the local output file path
	localFilePath := outputPath
	if localFilePath == "" {
		// If no output path is specified, use the original filename in the current directory
		_, fileName := filepath.Split(filePath)
		if fileName == "" {
			pan.PrintError(fmt.Sprintf("Error: Invalid file path: %s", filePath))
			os.Exit(1)
		}
		localFilePath = fileName
	}

	pan.PrintSuccess(fmt.Sprintf("Downloading file '%s' from Baidu Pan to '%s'...", filePath, localFilePath))

	err := client.DownloadFileToPath(filePath, localFilePath)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error downloading file: %v", err))
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("File downloaded successfully to: %s", localFilePath))
}

func uploadCommand(client *pan.Client) {
	uploadFlags := pflag.NewFlagSet("ul", pflag.ExitOnError)
	var localFilePath string
	var remoteFilePath string
	var help bool

	uploadFlags.StringVarP(&localFilePath, "source", "s", "", "Local file path to upload (required)")
	uploadFlags.StringVarP(&remoteFilePath, "destination", "d", "", "Remote file path in Baidu Pan (required, e.g., /path/to/your/file.txt)")
	uploadFlags.BoolVarP(&help, "help", "h", false, "Show help for upload command")

	if err := uploadFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		uploadFlags.PrintDefaults()
		return
	}

	if localFilePath == "" {
		pan.PrintError("Error: -f or --file flag is required to specify the local file to upload.")
		uploadFlags.PrintDefaults()
		os.Exit(1)
	}

	if remoteFilePath == "" {
		pan.PrintError("Error: -d or --dir flag is required to specify the remote file path.")
		uploadFlags.PrintDefaults()
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("Uploading local file '%s' to Baidu Pan as '%s'...", localFilePath, remoteFilePath))

	err := client.UploadFile(localFilePath, remoteFilePath)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error uploading file: %v", err))
		os.Exit(1)
	}

	_, fileName := filepath.Split(localFilePath)
	pan.PrintSuccess(fmt.Sprintf("File '%s' uploaded successfully to '%s'.", fileName, remoteFilePath))
}

func removeCommand(client *pan.Client) {
	removeFlags := pflag.NewFlagSet("rm", pflag.ExitOnError)
	var remotePath string
	var force bool
	var help bool

	removeFlags.StringVarP(&remotePath, "source", "s", "", "Remote file or directory path to remove (required)")
	removeFlags.BoolVarP(&force, "force", "y", false, "Force removal without confirmation")
	removeFlags.BoolVarP(&help, "help", "h", false, "Show help for remove command")

	if err := removeFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		removeFlags.PrintDefaults()
		return
	}

	if remotePath == "" {
		pan.PrintError("Error: -r or --remote-path flag is required to specify the file or directory to remove.")
		removeFlags.PrintDefaults()
		os.Exit(1)
	}

	// If not in force mode, ask for confirmation
	if !force {
		fmt.Printf("Are you sure you want to remove '%s'? This operation cannot be undone. (y/N): ", remotePath)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			pan.PrintSuccess("Remove operation cancelled.")
			return
		}
	}

	pan.PrintSuccess(fmt.Sprintf("Removing '%s' from Baidu Pan...", remotePath))

	err := client.RemoveFile(remotePath)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error removing file: %v", err))
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("'%s' removed successfully from Baidu Pan.", remotePath))
}

func moveCommand(client *pan.Client) {
	moveFlags := pflag.NewFlagSet("mv", pflag.ExitOnError)
	var sourcePath string
	var destPath string
	var force bool
	var help bool

	moveFlags.StringVarP(&sourcePath, "source", "s", "", "Source file or directory path to move (required)")
	moveFlags.StringVarP(&destPath, "destination", "d", "", "Destination directory path (required)")
	moveFlags.BoolVarP(&force, "force", "y", false, "Force move without confirmation")
	moveFlags.BoolVarP(&help, "help", "h", false, "Show help for move command")

	if err := moveFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		moveFlags.PrintDefaults()
		return
	}

	if sourcePath == "" {
		pan.PrintError("Error: -s or --source flag is required to specify the file or directory to move.")
		moveFlags.PrintDefaults()
		os.Exit(1)
	}

	if destPath == "" {
		pan.PrintError("Error: -d or --destination flag is required to specify the destination directory.")
		moveFlags.PrintDefaults()
		os.Exit(1)
	}

	// If not in force mode, ask for confirmation
	if !force {
		fmt.Printf("Are you sure you want to move '%s' to '%s'? (y/N): ", sourcePath, destPath)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			pan.PrintSuccess("Move operation cancelled.")
			return
		}
	}

	pan.PrintSuccess(fmt.Sprintf("Moving '%s' to '%s' in Baidu Pan...", sourcePath, destPath))

	err := client.MoveFile(sourcePath, destPath)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error moving file: %v", err))
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("'%s' moved successfully to '%s' in Baidu Pan.", sourcePath, destPath))
}

func renameCommand(client *pan.Client) {
	renameFlags := pflag.NewFlagSet("rn", pflag.ExitOnError)
	var sourcePath string
	var newName string
	var force bool
	var help bool

	renameFlags.StringVarP(&sourcePath, "source", "s", "", "Source file or directory path to rename (required)")
	renameFlags.StringVarP(&newName, "newname", "n", "", "New name for the file or directory (required)")
	renameFlags.BoolVarP(&force, "force", "y", false, "Force rename without confirmation")
	renameFlags.BoolVarP(&help, "help", "h", false, "Show help for rename command")

	if err := renameFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		renameFlags.PrintDefaults()
		return
	}

	if sourcePath == "" {
		pan.PrintError("Error: -s or --source flag is required to specify the file or directory to rename.")
		renameFlags.PrintDefaults()
		os.Exit(1)
	}

	if newName == "" {
		pan.PrintError("Error: -n or --newname flag is required to specify the new name.")
		renameFlags.PrintDefaults()
		os.Exit(1)
	}

	// Extract the parent directory from the source path to construct the new full path
	dir := filepath.Dir(sourcePath)
	if dir == "." {
		dir = "/"
	}
	newPath := filepath.Join(dir, newName)

	// If not in force mode, ask for confirmation
	if !force {
		fmt.Printf("Are you sure you want to rename '%s' to '%s'? (y/N): ", sourcePath, newPath)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			pan.PrintSuccess("Rename operation cancelled.")
			return
		}
	}

	pan.PrintSuccess(fmt.Sprintf("Renaming '%s' to '%s' in Baidu Pan...", sourcePath, newPath))

	err := client.RenameFile(sourcePath, newName)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error renaming file: %v", err))
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("'%s' renamed successfully to '%s' in Baidu Pan.", sourcePath, newPath))
}

func copyCommand(client *pan.Client) {
	copyFlags := pflag.NewFlagSet("cp", pflag.ExitOnError)
	var sourcePath string
	var destPath string
	var help bool

	copyFlags.StringVarP(&sourcePath, "source", "s", "", "Source file or directory path to copy (required)")
	copyFlags.StringVarP(&destPath, "destination", "d", "", "Destination file or directory path (required)")
	copyFlags.BoolVarP(&help, "help", "h", false, "Show help for copy command")

	if err := copyFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		copyFlags.PrintDefaults()
		return
	}

	if sourcePath == "" {
		pan.PrintError("Error: -s or --source flag is required to specify the source file or directory to copy.")
		copyFlags.PrintDefaults()
		os.Exit(1)
	}

	if destPath == "" {
		pan.PrintError("Error: -d or --destination flag is required to specify the destination path.")
		copyFlags.PrintDefaults()
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("Copying '%s' to '%s' in Baidu Pan...", sourcePath, destPath))

	err := client.CopyFile(sourcePath, destPath)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error copying file: %v", err))
		os.Exit(1)
	}
}

func mkdirCommand(client *pan.Client) {
	mkdirFlags := pflag.NewFlagSet("md", pflag.ExitOnError)
	var dirPath string
	var help bool

	mkdirFlags.StringVarP(&dirPath, "path", "p", "", "Directory path to create in Baidu Pan (required)")
	mkdirFlags.BoolVarP(&help, "help", "h", false, "Show help for mkdir command")

	if err := mkdirFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		mkdirFlags.PrintDefaults()
		return
	}

	if dirPath == "" {
		pan.PrintError("Error: -d or --dir flag is required to specify the directory path to create.")
		mkdirFlags.PrintDefaults()
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("Creating directory '%s' in Baidu Pan...", dirPath))

	err := client.CreateDir(dirPath)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error creating directory: %v", err))
		os.Exit(1)
	}
	pan.PrintSuccess(fmt.Sprintf("Directory '%s' created successfully.", dirPath))
}

func infoCommand(client *pan.Client) {
	infoFlags := pflag.NewFlagSet("if", pflag.ExitOnError)
	var filePath string
	var help bool

	infoFlags.StringVarP(&filePath, "path", "p", "", "File path in Baidu Pan to get information for (required)")
	infoFlags.BoolVarP(&help, "help", "h", false, "Show help for info command")

	if err := infoFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		infoFlags.PrintDefaults()
		return
	}

	if filePath == "" {
		pan.PrintError("Error: -f or --file flag is required to specify the file path to get information for.")
		infoFlags.PrintDefaults()
		os.Exit(1)
	}

	pan.PrintSuccess(fmt.Sprintf("Getting information for file: '%s' in Baidu Pan...", filePath))

	fileInfo, err := client.GetAndDisplayFileInfo(filePath)
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error getting file information: %v", err))
		os.Exit(1)
	}

	fmt.Print(pan.FormatFileInfo(fileInfo))
}

func diskInfoCommand(client *pan.Client) {
	diskInfoFlags := pflag.NewFlagSet("di", pflag.ExitOnError)
	var help bool

	diskInfoFlags.BoolVarP(&help, "help", "h", false, "Show help for disk info command")

	if err := diskInfoFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		diskInfoFlags.PrintDefaults()
		return
	}

	pan.PrintSuccess("Getting disk information from Baidu Pan...")

	diskInfo, err := client.GetDiskInfo()
	if err != nil {
		pan.PrintError(fmt.Sprintf("Error getting disk information: %v", err))
		os.Exit(1)
	}

	fmt.Print(pan.FormatDiskInfo(diskInfo))
}

func refreshTokenCommand(client *pan.Client) {
	refreshFlags := pflag.NewFlagSet("ar", pflag.ExitOnError)
	var help bool

	refreshFlags.BoolVarP(&help, "help", "h", false, "Show help for refresh token command")

	if err := refreshFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		refreshFlags.PrintDefaults()
		return
	}

	pan.PrintSuccess("Attempting to refresh access token...")

	// Check if there's a refresh token available
	if client.HasValidToken() {
		err := client.LoadTokens()
		if err != nil {
			pan.PrintError(fmt.Sprintf("Error loading existing tokens: %v", err))
			os.Exit(1)
		}

		if !client.HasRefreshToken() {
			pan.PrintError("No refresh token available, cannot refresh access token.")
			os.Exit(1)
		}

		err = client.RefreshToken()
		if err != nil {
			pan.PrintError(fmt.Sprintf("Error refreshing token: %v", err))
			os.Exit(1)
		}

		pan.PrintSuccess("Access token refreshed successfully and saved to .bdfs_certs")
	} else {
		pan.PrintError("No token file found, cannot refresh access token.")
		os.Exit(1)
	}
}

func versionCommand() {
	versionFlags := pflag.NewFlagSet("version", pflag.ExitOnError)
	var help bool

	versionFlags.BoolVarP(&help, "help", "h", false, "Show help for version command")

	if err := versionFlags.Parse(os.Args[2:]); err != nil {
		return
	}

	if help {
		versionFlags.PrintDefaults()
		return
	}

	fmt.Printf("go-bdfs version %s\n", VERSION)
}

func showHelp() {
	fmt.Println("go-bdfs: Baidu Pan client")
	fmt.Println("Usage: go-bdfs <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  ls          List files in a directory")
	fmt.Println("              Usage: go-bdfs ls -p <path>")
	fmt.Println("              Flags: -p, --path <path> (default: /)")
	fmt.Println("")
	fmt.Println("  dl          Download a file from Baidu Pan")
	fmt.Println("              Usage: go-bdfs dl -s <source> -d <destination>")
	fmt.Println("              Flags: -s, --source <source> (required), -d, --destination <destination> (optional)")
	fmt.Println("")
	fmt.Println("  ul          Upload a file to Baidu Pan")
	fmt.Println("              Usage: go-bdfs ul -s <source> -d <destination>")
	fmt.Println("              Flags: -s, --source <source> (required), -d, --destination <destination> (required)")
	fmt.Println("")
	fmt.Println("  rm          Remove a file or directory from Baidu Pan")
	fmt.Println("              Usage: go-bdfs rm -s <source> [-y]")
	fmt.Println("              Flags: -s, --source <source> (required), -y, --force (optional)")
	fmt.Println("")
	fmt.Println("  mv          Move a file or directory to another directory in Baidu Pan")
	fmt.Println("              Usage: go-bdfs mv -s <source> -d <destination> [-y]")
	fmt.Println("              Flags: -s, --source <source> (required), -d, --destination <destination> (required), -y, --force (optional)")
	fmt.Println("")
	fmt.Println("  rn          Rename a file or directory in Baidu Pan")
	fmt.Println("              Usage: go-bdfs rn -s <source> -n <newname>")
	fmt.Println("              Flags: -s, --source <source> (required), -n, --newname <newname> (required)")
	fmt.Println("")
	fmt.Println("  md          Create a directory in Baidu Pan")
	fmt.Println("              Usage: go-bdfs md -p <path>")
	fmt.Println("              Flags: -p, --path <path> (required)")
	fmt.Println("")
	fmt.Println("  cp          Copy a file or directory in Baidu Pan")
	fmt.Println("              Usage: go-bdfs cp -s <source> -d <destination>")
	fmt.Println("              Flags: -s, --source <source> (required), -d, --destination <destination> (required)")
	fmt.Println("")
	fmt.Println("  if          Get information about a file in Baidu Pan")
	fmt.Println("              Usage: go-bdfs if -p <path>")
	fmt.Println("              Flags: -p, --path <path> (required)")
	fmt.Println("")
	fmt.Println("  di          Get disk information (storage usage) from Baidu Pan")
	fmt.Println("              Usage: go-bdfs di")
	fmt.Println("              Flags: -h, --help (optional)")
	fmt.Println("")
	fmt.Println("  ar          Refresh the access token using the refresh token")
	fmt.Println("              Usage: go-bdfs ar")
	fmt.Println("              Flags: -h, --help (optional)")
	fmt.Println("")
	fmt.Println("  version     Show the version information")
	fmt.Println("              Usage: go-bdfs version")
	fmt.Println("              Flags: -h, --help (optional)")
	fmt.Println("")
	fmt.Println("  help        Show this help message")
	fmt.Println("")
	fmt.Println("Use 'go-bdfs <command> -h' or 'go-bdfs <command> --help' for more information about a command.")
}
