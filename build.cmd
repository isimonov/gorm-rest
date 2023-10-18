for %%a in (".") do set CURRENT_DIR_NAME=%%~na

if not exist "./build" mkdir ./build

set GOOS=linux
set GOARCH=amd64
go build -o ./build/%CURRENT_DIR_NAME% .

set GOOS=windows
set GOARCH=amd64
go build -o ./build/%CURRENT_DIR_NAME%.exe .