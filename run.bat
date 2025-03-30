goversioninfo -icon=library/assets/rof2plus.ico -manifest=rof2plus.exe.manifest -o=rsrc.syso versioninfo.json
go build -ldflags "-X main.Version=dev"
move rof2plus.exe bin/rof2plus.exe
cd bin
rof2plus.exe start test