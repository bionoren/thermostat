GOBIN=. GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 go build -ldflags="-s -w" thermostat
