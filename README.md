# go-bdfs

A command-line tool for managing files on Baidu Cloud Disk (Baidu Pan) using the official API.

## Overview

go-bdfs is a Go-based command-line client that allows you to interact with your Baidu Cloud Disk account. The tool supports various file operations such as listing, uploading, downloading, moving, copying, and deleting files and directories.

## Features

- List files and directories
- Upload files to Baidu Cloud Disk
- Download files from Baidu Cloud Disk
- Create directories
- Move, copy, rename, and delete files and directories
- View file information and disk usage
- Automatic token refresh
- Device code authorization flow

## Prerequisites

To use this tool, you need to register a Baidu Cloud Disk application to obtain a `client_id` and `client_secret`. You can register an application through the Baidu Open Platform.

## Installation

### Prerequisites

You need to have Go installed on your system. This project uses Go 1.25.4 or later.

### Building from source

1. Clone the repository:
```bash
git clone https://github.com/your-username/go-bdfs.git
cd go-bdfs
```

2. Build the project:
```bash
# Using the provided build script
./build.sh

# Or build directly with Go
go build -o build/go-bdfs
```

The binary will be available in the `build/` directory.

## Configuration

The tool requires your Baidu API credentials to work. You can configure them in two ways:

### Environment Variables

Set the following environment variables:

```bash
export BDU_PAN_CLIENT_ID="your_client_id"
export BDU_PAN_CLIENT_SECRET="your_client_secret"
```

### Configuration File

Create a configuration file at `~/.local/app/bdfs/config.toml`:

```toml
# Direct configuration
client_id = "your_client_id"
client_secret = "your_client_secret"
```

Or with the `[baidu]` table:

```toml
[baidu]
client_id = "your_client_id"
client_secret = "your_client_secret"
```

## Usage

### Authorization

On first use, the tool will prompt you to complete the device authorization flow:

```bash
./build/go-bdfs ls
```

You will be asked to visit a URL and enter an authorization code.

### Commands

The tool supports the following commands:

#### List Files (`ls`)

List files and directories in a specified path:

```bash
go-bdfs ls -p /path/to/directory
```

Options:
- `-p, --path`: Directory to list (default: `/`)

#### Download File (`dl`)

Download a file from Baidu Cloud Disk:

```bash
go-bdfs dl -s /remote/path/file.txt -d ./local/path/file.txt
```

Options:
- `-s, --source`: File path in Baidu Cloud Disk to download (required)
- `-d, --destination`: Local output file path (optional, defaults to current directory with original filename)

#### Upload File (`ul`)

Upload a file to Baidu Cloud Disk:

```bash
go-bdfs ul -s ./local/path/file.txt -d /remote/path/file.txt
```

Options:
- `-s, --source`: Local file path to upload (required)
- `-d, --destination`: Remote file path in Baidu Cloud Disk (required)

#### Remove File/Directory (`rm`)

Remove a file or directory from Baidu Cloud Disk:

```bash
go-bdfs rm -s /path/to/file/or/directory
```

Options:
- `-s, --source`: Remote file or directory path to remove (required)
- `-y, --force`: Force removal without confirmation

#### Move File/Directory (`mv`)

Move a file or directory to another directory in Baidu Cloud Disk:

```bash
go-bdfs mv -s /source/path -d /destination/directory
```

Options:
- `-s, --source`: Source file or directory path to move (required)
- `-d, --destination`: Destination directory path (required)
- `-y, --force`: Force move without confirmation

#### Rename File/Directory (`rn`)

Rename a file or directory in Baidu Cloud Disk:

```bash
go-bdfs rn -s /path/to/file -n new_name
```

Options:
- `-s, --source`: Source file or directory path to rename (required)
- `-n, --newname`: New name for the file or directory (required)

#### Create Directory (`md`)

Create a directory in Baidu Cloud Disk:

```bash
go-bdfs md -p /path/to/new/directory
```

Options:
- `-p, --path`: Directory path to create in Baidu Cloud Disk (required)

#### Copy File/Directory (`cp`)

Copy a file or directory in Baidu Cloud Disk:

```bash
go-bdfs cp -s /source/path -d /destination/path
```

Options:
- `-s, --source`: Source file or directory path to copy (required)
- `-d, --destination`: Destination file or directory path (required)

#### File Information (`if`)

Get information about a file in Baidu Cloud Disk:

```bash
go-bdfs if -p /path/to/file
```

Options:
- `-p, --path`: File path in Baidu Cloud Disk to get information for (required)

#### Disk Information (`di`)

Get disk usage information from Baidu Cloud Disk:

```bash
go-bdfs di
```

No options required.

#### Access Token Refresh (`ar`)

Refresh the access token using the refresh token:

```bash
go-bdfs ar
```

No options required.

### Help

To see all available commands and options:

```bash
go-bdfs
go-bdfs <command> -h  # for specific command help
```

## Authentication and Token Storage

The tool automatically handles the OAuth 2.0 device authorization flow and stores the access token in `~/.local/app/bdfs/.bdfs_certs`. The tool also automatically refreshes the token when it expires.

## Contributing

Feel free to submit issues and enhancement requests. Pull requests are welcome.

## License

This project is licensed under the MIT License - see the LICENSE file for details.