package system

import (
	"github.com/sirupsen/logrus"
	"math"
	"periph.io/x/periph/host"
)

func init() {
	if _, err := host.Init(); err != nil {
		logrus.WithError(err).Panic("Could not init periph host")
	}
}

func FarenheitFromCelcius(celcius float64) float64 {
	return celcius*1.8 + 32
}

// heatIndex calculates the heat index from the given temperature and humidity
//
// https://www.wpc.ncep.noaa.gov/html/heatindex_equation.shtml
// The computation of the heat index is a refinement of a result obtained by multiple regression analysis carried out by Lans P. Rothfusz and described in a 1990 National Weather Service (NWS) Technical Attachment (SR 90-23).  The regression equation of Rothfusz is
// HI = -42.379 + 2.04901523*T + 10.14333127*RH - .22475541*T*RH - .00683783*T*T - .05481717*RH*RH + .00122874*T*T*RH + .00085282*T*RH*RH - .00000199*T*T*RH*RH
// where T is temperature in degrees F and RH is relative humidity in percent.  HI is the heat index expressed as an apparent temperature in degrees F.  If the RH is less than 13% and the temperature is between 80 and 112 degrees F, then the following adjustment is subtracted from HI:
// ADJUSTMENT = [(13-RH)/4]*SQRT{[17-ABS(T-95.)]/17}
// where ABS and SQRT are the absolute value and square root functions, respectively.  On the other hand, if the RH is greater than 85% and the temperature is between 80 and 87 degrees F, then the following adjustment is added to HI:
// ADJUSTMENT = [(RH-85)/10] * [(87-T)/5]
// The Rothfusz regression is not appropriate when conditions of temperature and humidity warrant a heat index value below about 80 degrees F. In those cases, a simpler formula is applied to calculate values consistent with Steadman's results:
// HI = 0.5 * {T + 61.0 + [(T-68.0)*1.2] + (RH*0.094)}
// In practice, the simple formula is computed first and the result averaged with the temperature. If this heat index value is 80 degrees F or higher, the full regression equation along with any adjustment as described above is applied.
// The Rothfusz regression is not valid for extreme temperature and relative humidity conditions beyond the range of data considered by Steadman.
func HeatIndex(temp, hum float64) float64 {
	var hi float64
	if temp < 80 {
		temp = 0.5 * (temp + 61.0 + (temp-68.0)*1.2 + hum*0.094)
		hi = temp
	}
	if temp >= 80 {
		t2 := temp * temp
		h2 := hum * hum
		hi = -42.379 + 2.04901523*temp + 10.14333127*hum - .22475541*temp*hum - .00683783*t2 - .05481717*h2 + .00122874*t2*hum + .00085282*temp*h2 - .00000199*t2*h2
	}
	if hum < 13 && temp >= 80 && temp <= 112 {
		abs := temp - 95
		if abs < 0 {
			abs = -abs
		}
		hi -= (13 - hum) * math.Sqrt((17-abs)/17) / 4
	}
	if hum > 85 && temp > 80 && temp < 87 {
		hi += (hum - 85) / 10 * (87 - temp) / 5
	}

	return hi
}
