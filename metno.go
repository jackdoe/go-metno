package metno

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// used https://mholt.github.io/json-to-go/ to generate it from the
// json output of api.met.no, hats off to @mholt for doing this
// awesome software, and also to the api.met.no guys who provide their
// data for free and have awesome service
type MetNoWeatherOutput struct {
	Product *struct {
		Class string `json:"class"`
		Time  []*struct {
			To       time.Time `json:"to"`
			Datatype string    `json:"datatype"`
			From     time.Time `json:"from"`
			Location *struct {
				Longitude float64 `json:"longitude,string"`
				Altitude  float64 `json:"altitude,string"`
				Latitude  float64 `json:"latitude,string"`

				Fog *struct {
					ID      string  `json:"id"`
					Percent float64 `json:"percent,string"`
				} `json:"fog"`
				TemperatureProbability *struct {
					Unit  string  `json:"unit"`
					Value float64 `json:"value,string"`
				} `json:"temperatureProbability"`
				WindProbability *struct {
					Unit  string  `json:"unit"`
					Value float64 `json:"value,string"`
				} `json:"windProbability"`
				Pressure *struct {
					ID    string  `json:"id"`
					Unit  string  `json:"unit"`
					Value float64 `json:"value,string"`
				} `json:"pressure"`
				Cloudiness *struct {
					Percent float64 `json:"percent,string"`
					ID      string  `json:"id"`
				} `json:"cloudiness"`
				WindDirection *struct {
					Deg  float64 `json:"deg,string"`
					Name string  `json:"name"`
					ID   string  `json:"id"`
				} `json:"windDirection"`

				DewpointTemperature *struct {
					ID    string  `json:"id"`
					Value float64 `json:"value,string"`
					Unit  string  `json:"unit"`
				} `json:"dewpointTemperature"`
				WindGust *struct {
					Mps float64 `json:"mps,string"`
					ID  string  `json:"id"`
				} `json:"windGust"`
				Humidity *struct {
					Value float64 `json:"value,string"`
					Unit  string  `json:"unit"`
				} `json:"humidity"`
				AreaMaxWindSpeed *struct {
					Mps float64 `json:"mps,string"`
				} `json:"areaMaxWindSpeed"`
				WindSpeed *struct {
					Beaufort string  `json:"beaufort"`
					ID       string  `json:"id"`
					Name     string  `json:"name"`
					Mps      float64 `json:"mps,string"`
				} `json:"windSpeed"`
				Temperature *struct {
					Value float64 `json:"value,string"`
					Unit  string  `json:"unit"`
					ID    string  `json:"id"`
				} `json:"temperature"`
				LowClouds *struct {
					Percent float64 `json:"percent,string"`
					ID      string  `json:"id"`
				} `json:"lowClouds"`
				MediumClouds *struct {
					Percent float64 `json:"percent,string"`
					ID      string  `json:"id"`
				} `json:"mediumClouds"`
				HighClouds *struct {
					ID      string  `json:"id"`
					Percent float64 `json:"percent,string"`
				} `json:"highClouds"`
			} `json:"location"`
		} `json:"time"`
	} `json:"product"`
	Created time.Time `json:"created"`
}

// Query for https://api.met.no/weatherapi/locationforecast/1.9/documentation
// Those guys are the best, please do not query them a lot and read the docs and https://api.met.no/license_data.html
// Example:
//	client := SimpleClient(1) // returns http.Client with timeout of 1 second
//	out, err := LocationForecast(client, 60.1, 8.0, 10) // latitude, longitude, meters above sea level (used only outside of Norway)
//
//	for _, v := range out.Product.Time {
//		if v.Location.Temperature != nil {
//			log.Printf("%s temp: %.2f %s\n", v.From, v.Location.Temperature.Value, v.Location.Temperature.Unit)
//		}
//	}
// outputs:
//  2018-08-25 15:00:00 +0000 UTC temp: 7.80 celsius
//  2018-08-25 16:00:00 +0000 UTC temp: 4.00 celsius
//  2018-08-25 17:00:00 +0000 UTC temp: 5.50 celsius
//  2018-08-25 18:00:00 +0000 UTC temp: 4.40 celsius
//  2018-08-25 19:00:00 +0000 UTC temp: 2.50 celsius
//  2018-08-25 20:00:00 +0000 UTC temp: 2.10 celsius
//  2018-08-25 21:00:00 +0000 UTC temp: 1.20 celsius
//  2018-08-25 22:00:00 +0000 UTC temp: 0.90 celsius
//  2018-08-25 23:00:00 +0000 UTC temp: 0.70 celsius
//  2018-08-26 00:00:00 +0000 UTC temp: 1.30 celsius
//  2018-08-26 01:00:00 +0000 UTC temp: 1.60 celsius
//  ...
func LocationForecast(client *http.Client, lat, lng float64, msl int) (*MetNoWeatherOutput, error) {
	req, err := http.NewRequest("GET", "https://api.met.no/weatherapi/locationforecast/1.9/", nil)

	q := req.URL.Query()
	q.Add("lat", fmt.Sprintf("%.6f", lat))
	q.Add("lon", fmt.Sprintf("%.6f", lng))
	q.Add("msl", fmt.Sprintf("%d", msl))
	req.URL.RawQuery = q.Encode()

	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("User-Agent", "go-metno client (https://github.com/jackdoe/go-metno)")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	out := MetNoWeatherOutput{}
	err = json.Unmarshal(buf, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// creates simple http client with timeout, in case you dont have your own
func SimpleClient(timeoutSeconds time.Duration) *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: timeoutSeconds * time.Second,
		}).Dial,
		DisableCompression:  false, // default is false, but make it explicit that gzip is handled from http.Transport
		TLSHandshakeTimeout: timeoutSeconds * time.Second,
	}

	proxyUrl, err := url.Parse(os.Getenv("HTTPS_PROXY"))
	if err == nil {
		netTransport.Proxy = http.ProxyURL(proxyUrl)
	}
	var netClient = &http.Client{
		Timeout:   time.Second * timeoutSeconds,
		Transport: netTransport,
	}

	return netClient
}
