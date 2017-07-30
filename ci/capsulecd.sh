#!/usr/bin/env bash
# retrieve the latest capsulecd release info
asset_url=$(curl -s https://api.github.com/repos/AnalogJ/capsulecd/releases/latest \
	| grep browser_download_url | grep 'capsulecd-linux' | cut -d '"' -f 4)

# download the capsulecd asset here.
curl -L -o capsulecd $asset_url

# make capsulecd executable
chmod +x capsulecd