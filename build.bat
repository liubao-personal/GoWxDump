go mod tidy
set CGO_ENABLED=1
set GOARCH=386
go build -x
copy GoWxDump.exe Release\GoWxDump.exe