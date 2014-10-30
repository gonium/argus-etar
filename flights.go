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
	NumMessages  int
	Tick         int
	Published    bool
}

type Flights map[string]Flight

func (f Flight) IsETAR() (retval bool) {
	retval = false
	if (f.Altitude < 3000 && f.Altitude > 0) &&
		(f.NumMessages > 10) && (f.VerticalRate < 0.0) {
		retval = true
	}
	return retval
}

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
		tmp.NumMessages = tmp.NumMessages + 1
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
		tmp.NumMessages = tmp.NumMessages + 1
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
		tmp.NumMessages = tmp.NumMessages + 1
		(*a)[hexident] = tmp
	}
}

func (a Flights) Tick() {
	for key, _ := range a {
		tmp := a[key]
		tmp.Tick = tmp.Tick + 1
		a[key] = tmp
		if time.Now().Sub(a[key].LastSeen) > (15 * time.Minute) {
			fmt.Println("Not seen for 15 minutes, deleting ", key)
			delete(a, key)
		}
	}
}

func (a Flights) String() string {
	buf := new(bytes.Buffer)
	w := new(tabwriter.Writer)
	w.Init(buf, 8, 0, 1, ' ', tabwriter.AlignRight)

	fmt.Fprintf(w, "ID\tCS\tAlt\tLat\tLon\tGS\tVV\tMsg\t\ttick\n")
	for key, value := range a {
		fmt.Fprintf(w, "%s\t%s\t%d\t%f\t%f\t%d\t%d\t%d\t\t%d\n",
			key, value.Callsign, value.Altitude, value.Latitude,
			value.Longitude, value.GroundSpeed, value.VerticalRate,
			value.NumMessages, value.Tick)
	}
	w.Flush()
	return buf.String()
}
