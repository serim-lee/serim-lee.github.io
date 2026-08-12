@echo off
setlocal
call "%~dp0run-go.cmd" run ./cmd/citations %*
exit /b %errorlevel%
