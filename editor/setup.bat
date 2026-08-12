@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "EDITOR_DIR=%~dp0"
set "RUNTIME_DIR=%EDITOR_DIR%runtime"
set "GOROOT=%RUNTIME_DIR%\go"
set "GOPATH=%RUNTIME_DIR%\gopath"
set "GOMODCACHE=%RUNTIME_DIR%\gomodcache"
set "GOCACHE=%RUNTIME_DIR%\gocache"
set "GOBIN=%RUNTIME_DIR%\bin"
set "PATH=%GOROOT%\bin;%GOBIN%;%PATH%"

if not exist "%RUNTIME_DIR%" mkdir "%RUNTIME_DIR%"

rem Keep Go/Wails project scans from descending into the local toolchain cache.
> "%RUNTIME_DIR%\go.mod" echo module editor-local-runtime
>> "%RUNTIME_DIR%\go.mod" echo.
>> "%RUNTIME_DIR%\go.mod" echo go 1.23

if not exist "%RUNTIME_DIR%\.portable-go-ready" (
  echo Downloading a portable Go toolchain into editor\runtime...
  powershell -NoProfile -ExecutionPolicy Bypass -File "%EDITOR_DIR%scripts\install-go.ps1" -RuntimeDir "%RUNTIME_DIR%"
  if errorlevel 1 exit /b !errorlevel!
)

echo Using:
go version
if errorlevel 1 exit /b %errorlevel%

echo.
echo Installing the Wails CLI into editor\runtime...
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
if errorlevel 1 exit /b %errorlevel%

echo.
echo Downloading project Go modules into editor\runtime...
pushd "%EDITOR_DIR%cmd\editor"
go mod download
set "RESULT=%errorlevel%"
popd
if not "%RESULT%"=="0" exit /b %RESULT%

echo.
echo Setup completed. All local toolchain files are under editor\runtime.
exit /b 0
