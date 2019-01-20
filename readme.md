Config:
* minTemp (int): default 60
* maxtemp (int): default 85
* apiPort (int): default 441
* apiCert (string): 
* apiKey (string): private key for the api certificate
* apiSecret (string): base64 encoded api secret key
* tempSensor (string): temperature sensor i2c bus address (hex in the form 0x##)
* fanPin (int): GPIO pin for the blower
* acPin (int): GPIO pin for the AC compressor
* heatPin (int): GPIO pin for the heater
* db.file (string): path to the database file
* log.type (string): stderr or file
* log.level (string): panic, error, warn, info, or debug
* log.file (string): filename (if log type is file)
* log.report (string): optional filename to store an activity log as a CSV file

Generate api secret
-------------------
dd if=/dev/urandom -bs=1 -count=128 | base64

Time
----
All time values are stored and computed in the local timezone. APIs transmit time data in UTC

Install
-------
Requires package: i2c-tools libi2c-dev

You now need to edit the modules conf file /etc/modules
Add these two lines:

```
i2c-dev
i2c-bcm2708
```

Update /boot/config.txt
Add to the bottom:

```
dtparam=i2c_arm=on
dtparam=i2c1=on
```
