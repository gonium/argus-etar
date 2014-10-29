package argus

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"
)

type Flight struct {
	Callsign     string
	Altitude     int
	GroundSpeed  int
	VerticalRate int
	Latitude     float64
	Longitude    float64
	FirstSeen    time.Time
	LastSeen     time.Time
	Published    bool
}

type Flights map[string]Flight

func (a *Flights) AddCallsign(hexident string, recvtime time.Time, callsign string) {
	_, ok := (*a)[hexident]
	if !ok { // We haven't seen this hexident before
		(*a)[hexident] = Flight{
			Callsign:  strings.TrimSpace(callsign),
			FirstSeen: recvtime,
			LastSeen:  recvtime,
		}
	} else {
		tmp := (*a)[hexident]
		tmp.Callsign = strings.TrimSpace(callsign)
		tmp.LastSeen = recvtime
		(*a)[hexident] = tmp
	}
}

func (a *Flights) AddPosition(hexident string, recvtime time.Time,
	altitude int, lat float64, lon float64) {
	_, ok := (*a)[hexident]
	if !ok { // We haven't seen this hexident before
		(*a)[hexident] = Flight{
			Callsign:  "       ",
			FirstSeen: recvtime,
			LastSeen:  recvtime,
			Altitude:  altitude,
			Latitude:  lat,
			Longitude: lon,
		}
	} else {
		tmp := (*a)[hexident]
		tmp.LastSeen = recvtime
		tmp.Altitude = altitude
		tmp.Latitude = lat
		tmp.Longitude = lon
		(*a)[hexident] = tmp
	}
}

func (a *Flights) AddVelocity(hexident string, recvtime time.Time,
	groundspeed int, verticalrate int) {
	_, ok := (*a)[hexident]
	if !ok { // We haven't seen this hexident before
		(*a)[hexident] = Flight{
			Callsign:     "       ",
			FirstSeen:    recvtime,
			LastSeen:     recvtime,
			GroundSpeed:  groundspeed,
			VerticalRate: verticalrate,
		}
	} else {
		tmp := (*a)[hexident]
		tmp.LastSeen = recvtime
		tmp.GroundSpeed = groundspeed
		tmp.VerticalRate = verticalrate
		(*a)[hexident] = tmp
	}
}

func (a Flights) String() string {
	buf := new(bytes.Buffer)
	w := new(tabwriter.Writer)
	w.Init(buf, 8, 0, 1, ' ', tabwriter.AlignRight)

	fmt.Fprintf(w, "ID\tCS\tAlt\tLat\tLon\tGS\t\tVV\n")
	for key, value := range a {
		fmt.Fprintf(w, "%s\t%s\t%d\t%f\t%f\t%d\t\t%d\n",
			key, value.Callsign, value.Altitude, value.Latitude,
			value.Longitude, value.GroundSpeed, value.VerticalRate)
	}
	w.Flush()
	return buf.String()
}
