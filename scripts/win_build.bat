@echo off

echo Building CrunchyCleaner for Windows amd64...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -o ..\crunchycleaner.exe ..\src\crunchycleaner.go
if errorlevel 1 (
    echo Windows build failed
    exit /b 1
    pause
)

echo Building CrunchyCleaner for Linux amd64...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o ..\crunchycleaner ..\src\crunchycleaner.go
if errorlevel 1 (
    echo Linux build failed
    exit /b 1
    pause
)

echo Build finished successfully.
timeout /t 3
exit
