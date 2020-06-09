GOBIN=. GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 go build -ldflags="-s -w" thermostat
cp website/* /usr/share/thermostat/
cp db/migrations/* /usr/share/thermostat/

#sudo cp thermostat.service /lib/systemd/system/thermostat.service
#sudo chmod 644 /lib/systemd/system/thermostat.service
#sudo chown root:root /lib/systemd/system/thermostat.service
