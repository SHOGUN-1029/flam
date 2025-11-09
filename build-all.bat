@echo off
echo ===========================================================
echo 🚀 Building cross-platform binaries for QueueCTL
echo ===========================================================

REM Disable CGO for static builds
set CGO_ENABLED=0

REM Create output directories
if not exist dist mkdir dist
if not exist dist\darwin mkdir dist\darwin
if not exist dist\linux mkdir dist\linux
if not exist dist\windows mkdir dist\windows
if not exist dist\freebsd mkdir dist\freebsd

REM ===========================
REM macOS Builds
REM ===========================
echo 🧱 Building for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -o dist\darwin\queuectl-darwin-amd64
set GOARCH=arm64
go build -o dist\darwin\queuectl-darwin-arm64

REM ===========================
REM Linux Builds
REM ===========================
echo 🧱 Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -o dist\linux\queuectl-linux-amd64
set GOARCH=386
go build -o dist\linux\queuectl-linux-386
set GOARCH=arm
go build -o dist\linux\queuectl-linux-arm
set GOARCH=arm64
go build -o dist\linux\queuectl-linux-arm64

REM ===========================
REM Windows Builds
REM ===========================
echo 🧱 Building for Windows...
set GOOS=windows
set GOARCH=amd64
go build -o dist\windows\queuectl-windows-amd64.exe
set GOARCH=386
go build -o dist\windows\queuectl-windows-386.exe

REM ===========================
REM FreeBSD Builds
REM ===========================
echo 🧱 Building for FreeBSD...
set GOOS=freebsd
set GOARCH=amd64
go build -o dist\freebsd\queuectl-freebsd-amd64
set GOARCH=386
go build -o dist\freebsd\queuectl-freebsd-386
set GOARCH=arm
go build -o dist\freebsd\queuectl-freebsd-arm

echo ===========================================================
echo ✅ All builds completed successfully!
echo 📦 Binaries are available under the 'dist\' directory.
echo ===========================================================
