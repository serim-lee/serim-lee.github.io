@echo off
setlocal
call "%~dp0run-go.cmd" run ./cmd/site validate
if errorlevel 1 exit /b %errorlevel%
echo.
echo Validation completed.
if not defined CI pause
