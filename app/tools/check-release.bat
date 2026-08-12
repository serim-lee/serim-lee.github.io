@echo off
setlocal EnableExtensions
call "%~dp0run-go.cmd" run ./cmd/releases check %*
exit /b %errorlevel%
