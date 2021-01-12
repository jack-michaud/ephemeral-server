build/ephemeralbot:
	mkdir -p build
	CGO_ENABLED=0 go build \
							-ldflags '-extldflags "-static"' \
							-o build/ephemeralbot \
							./bot/cmd/run/main.go

clean:
	rm -r build/*

