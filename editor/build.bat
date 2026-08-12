@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "EDITOR_DIR=%~dp0"
set "REPO_DIR=%EDITOR_DIR%.."
set "RUNTIME_DIR=%EDITOR_DIR%runtime"
set "GOROOT=%RUNTIME_DIR%\go"
set "GOPATH=%RUNTIME_DIR%\gopath"
set "GOMODCACHE=%RUNTIME_DIR%\gomodcache"
set "GOCACHE=%RUNTIME_DIR%\gocache"
set "GOBIN=%RUNTIME_DIR%\bin"
set "PATH=%GOROOT%\bin;%GOBIN%;%PATH%"
set "OUTPUT_NAME=Profile-Editor.exe"

if not exist "%RUNTIME_DIR%\.portable-go-ready" goto :not_setup
if not exist "%GOROOT%\bin\go.exe" goto :not_setup
if not exist "%GOBIN%\wails.exe" goto :not_setup

set "RELEASE_VERSION="
for /f "delims=" %%V in ('powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "%EDITOR_DIR%scripts\read-release-version.ps1" -RepositoryRoot "%REPO_DIR%"') do set "RELEASE_VERSION=%%V"
if not defined RELEASE_VERSION (
  echo ERROR: Could not read the program release version.
  exit /b 1
)

copy /Y "%REPO_DIR%\app\assets\js\publication-bibliography.js" "%EDITOR_DIR%cmd\editor\frontend\dist\publication-bibliography.js" >nul
if errorlevel 1 exit /b 1
copy /Y "%REPO_DIR%\app\assets\js\color-themes.js" "%EDITOR_DIR%cmd\editor\frontend\dist\color-themes.js" >nul
if errorlevel 1 exit /b 1
copy /Y "%REPO_DIR%\app\assets\js\data-version.js" "%EDITOR_DIR%cmd\editor\frontend\dist\data-version.js" >nul
if errorlevel 1 exit /b 1

pushd "%EDITOR_DIR%cmd\editor"
"%GOBIN%\wails.exe" build -clean -o "%OUTPUT_NAME%" -ldflags "-X main.compiledVersion=%RELEASE_VERSION%"
set "RESULT=%errorlevel%"
if not "!RESULT!"=="0" (
  popd
  exit /b !RESULT!
)

copy /Y "build\bin\%OUTPUT_NAME%" "%REPO_DIR%\%OUTPUT_NAME%" >nul
set "RESULT=%errorlevel%"
popd
if not "%RESULT%"=="0" exit /b %RESULT%

echo.
echo Build completed: %REPO_DIR%\%OUTPUT_NAME%
echo This root-level executable is intentionally tracked by Git.
exit /b 0

:not_setup
echo ERROR: The local editor toolchain is not installed.
echo Run editor\setup.bat first.
exit /b 1
