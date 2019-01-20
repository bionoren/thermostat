package main

import (
	"crypto/tls"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadApiCert(t *testing.T) {
	t.Parallel()

	viper.Set("apiCert", `-----BEGIN CERTIFICATE-----
MIICLTCCAZYCCQCHpcaq4CZ54TANBgkqhkiG9w0BAQsFADBbMQswCQYDVQQGEwJV
UzEOMAwGA1UECAwFVGV4YXMxDzANBgNVBAcMBkRhbGxhczEXMBUGA1UECgwOTGxh
bWEgU29mdHdhcmUxEjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0xOTAxMDcxNzE1MTla
Fw0zODAzMDgxNzE1MTlaMFsxCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVUZXhhczEP
MA0GA1UEBwwGRGFsbGFzMRcwFQYDVQQKDA5MbGFtYSBTb2Z0d2FyZTESMBAGA1UE
AwwJbG9jYWxob3N0MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDKpZQp7hy5
mOxDrw+DCZ27OKFiizLXmjFwLmCDRezThIUhKuHLoxd7miFaS6rxuq/Z6ltrHuim
OWi6cZLSQd4JeRxGRagEx0c5aV8helWKhqM1QfsSdkEmo1ms8XUZcHW/kvtd61w4
tlasWKQG6sm2+FTB7daCiZ3kzIvPWx0zTwIDAQABMA0GCSqGSIb3DQEBCwUAA4GB
AKpELbS84Ego7ENK2+ikqQv2FVmDsJp6yUQrgUCba+Vo+V6wer0reQtNgbuDLxs9
3uZdzTfSQvujvqW+kTbYFc4vN/VxjclWzzlbB+ewsoQSHFTSN5MzShfXUiNB2sg1
XY6RTBq2ecp1GqntcaIoIzhTCeT6o+fMUAl8OOonkpmz
-----END CERTIFICATE-----`)

	viper.Set("apiKey", `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDKpZQp7hy5mOxDrw+DCZ27OKFiizLXmjFwLmCDRezThIUhKuHL
oxd7miFaS6rxuq/Z6ltrHuimOWi6cZLSQd4JeRxGRagEx0c5aV8helWKhqM1QfsS
dkEmo1ms8XUZcHW/kvtd61w4tlasWKQG6sm2+FTB7daCiZ3kzIvPWx0zTwIDAQAB
AoGAHMpffYGN5TR7xLX3bzeLiFDoZNa/92+5vGVqYtwpZHe8blToVYUrTe089dYw
SD2sxDoOmO6AQTWA0pRWNrcS82a5jYWVI3O8BQDtYrry1yycvXMUaS0f2Z/mM6ZS
GmuW0+EDqAcLhMXGnOOkC67NRfMlNrpCDZBrU3pncXhIGakCQQDz4c/3meJ6gtiE
EAeuFO3smgA7hezCnsQSx0VK9/Ri+Ezz2G3sPIAaxeyBeNarkV8om/p5CaaA5hYp
fnnVuxmbAkEA1LdAk8NEEQJGh7UdQMH/LxzdKH/c79ijZ0CtMlwFakbUwNt6fm9u
qtBvfricTdIgNNBmeH2SFrc0gs2pHIgSXQJAFvPxtsPs5MrbxdIcZu3hVptH2lJI
biizG3FVvDCJ96aW13xPHCS1ic+G6siMq6kK46+Ka0nVOdxtyYn1vX/WcQJADk6l
EUs48MvuYoJUDV7/AvQ2C9tNyPQRSYiYHaMC2jsZZD9e5dIo52RNm4BfQvy3HdZG
jiQkB1MbPREIJtsgIQJBAL+ORx9ODd3Dz1VmR6CuXBX6UnCYHnksgZZUyQX9ez5k
KIztiUEnd9QCL6HahOLxcCDH5HwXF7I7SYE3fuS3kro=
-----END RSA PRIVATE KEY-----`)

	cert, key, err := loadApiCert()
	require.NoError(t, err)
	assert.NotEmpty(t, cert)
	assert.NotEmpty(t, key)

	certificate, err := tls.X509KeyPair(cert, key)
	require.NoError(t, err)
	assert.Len(t, certificate.Certificate, 1)
}
