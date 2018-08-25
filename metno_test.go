package metno

import (
	"testing"
)

func TestRemote(t *testing.T) {
	client := SimpleClient(1)
	out, err := LocationForecast(client, 60, 8, 0)

	for _, v := range out.Product.Time {
		if v.Location.Temperature != nil {
			t.Logf("%s temp: %.2f %s\n", v.From, v.Location.Temperature.Value, v.Location.Temperature.Unit)
		}
	}

	if err != nil {
		t.FailNow()
	}
}
