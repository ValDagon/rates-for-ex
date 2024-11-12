APP_NAME="R-F-E"

build:
    go build -ldflags "-X main.version=$(VERSION)" -o $(APP_NAME)

clean:
    rm -f $(APP_NAME)

.PHONY: build clean
