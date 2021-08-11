#!/bin/sh

REPO_NAME=frzleaf/aolang
EXECUTION_FILE=aolang_server_linux


get_latest_release() {
  curl --silent "https://api.github.com/repos/${REPO_NAME}/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                                        # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                                # Pluck JSON value
}

latest_version=$(get_latest_release)

cd /tmp/
wget "https://github.com/${REPO_NAME}/releases/download/${latest_version}/${EXECUTION_FILE}"

clear
chmod +x aolang_server_linux
./aolang_server_linux :9999
