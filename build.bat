
goversioninfo -icon=assets/rof2plus.ico -manifest=rof2plus.exe.manifest -o=rsrc.syso versioninfo.json
go build -ldflags "-s -w -X main.Version=0.1.1.2" || exit /b
move rof2plus.exe bin/rof2plus.exe || exit /b
cd bin || exit /b