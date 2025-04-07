run:
	@mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/rof2plus .
	cd bin && ./rof2plus start