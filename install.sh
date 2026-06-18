#!/bin/bash
set -e

echo "🚀 Installing MacRemote via Homebrew..."

# 1. Tap the repository explicitly
echo "👉 Tapping adarsh9780/mac_remote..."
brew tap adarsh9780/mac_remote https://github.com/adarsh9780/mac_remote

# 2. Trust the custom tap (Homebrew requires this for third-party taps)
echo "🔐 Trusting the custom tap..."
brew trust adarsh9780/mac_remote

# 3. Install the application
echo "📦 Downloading and installing MacRemote..."
brew install macremote

echo "✅ Installation complete! You can find MacRemote in your /Applications folder."
