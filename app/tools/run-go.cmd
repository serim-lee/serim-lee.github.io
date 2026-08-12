@echo off
setlocal EnableExtensions
set "REPO_DIR=%~dp0..\.."
set "RUNTIME_DIR=%REPO_DIR%\editor\runtime"
set "GO_TOOL=go.exe"
if exist "%RUNTIME_DIR%\go\bin\go.exe" (
  set "GO_TOOL=%RUNTIME_DIR%\go\bin\go.exe"
  set "GOROOT=%RUNTIME_DIR%\go"
  set "GOPATH=%RUNTIME_DIR%\gopath"
  set "GOMODCACHE=%RUNTIME_DIR%\gomodcache"
  set "GOCACHE=%RUNTIME_DIR%\gocache"
) else (
  where go.exe >nul 2>&1
  if errorlevel 1 (
    echo ERROR: Go was not found. Run editor\setup.bat or install Go.
    exit /b 1
  )
)
pushd "%REPO_DIR%\editor"
if errorlevel 1 exit /b 1
"%GO_TOOL%" %*
set "RESULT=%errorlevel%"
popd
exit /b %RESULT%
