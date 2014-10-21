package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
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

var argus Flights

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

func init() {
	// Initialize our flight surveillance recorder data structure
	argus = make(map[string]Flight)
}

func main() {
	// TODO: Take this from the command line
	conn, err := net.Dial("tcp", "192.168.1.22:30003")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot connect to SBS stream:", err.Error())
		os.Exit(1)
	}
	connbuf := bufio.NewReader(conn)
	cnt := 0
	for {
		str, err := connbuf.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to read line from SBS stream:", err.Error())
			os.Exit(1)
		}
		if len(str) > 0 {
			//fmt.Println(str)
			sbs := strings.Split(str, ",")
			if len(sbs) != 22 {
				fmt.Fprintln(os.Stderr, "Received SBS message w/ invalid field count:", len(sbs))
				continue
			}
			// We can now parse the SBS line - see
			// http://www.homepages.mcb.net/bones/SBS/Article/Barebones42_Socket_Data.htm
			// For now, only parse MSG messages.
			if sbs[0] == "MSG" {
				cnt += 1
				if cnt == 300 {
					fmt.Println(argus)
					cnt = 0
				}
				recvTime := time.Now()
				hexIdent := sbs[4]
				switch sbs[1] { // This contains the subtype
				case "1": // ES Identification
					//					fmt.Println("Hexident:", hexIdent, "Callsign:", sbs[10])
					argus.AddCallsign(hexIdent, recvTime, sbs[10])
				case "3": // ES Airborne Position Message
					//fmt.Println("Hexident:", hexIdent, "Altitude:", sbs[11], "Lat:",
					//	sbs[14], "Lon:", sbs[15])
					alt, err := strconv.Atoi(sbs[11])
					if err != nil {
						continue
					}
					lat, err2 := strconv.ParseFloat(sbs[14], 64)
					if err2 != nil {
						continue
					}
					lon, err3 := strconv.ParseFloat(sbs[15], 64)
					if err3 != nil {
						continue
					}
					argus.AddPosition(hexIdent, recvTime, alt, lat, lon)
				case "4": // ES Airborne Velocity Message
					//fmt.Println("Hexident:", hexIdent, "Ground Speed:", sbs[12],
					//	"Vertical Rate:", sbs[16])
					groundspeed, err := strconv.Atoi(sbs[12])
					if err != nil {
						continue
					}
					verticalrate, err := strconv.Atoi(sbs[16])
					if err != nil {
						continue
					}
					argus.AddVelocity(hexIdent, recvTime, groundspeed,
						verticalrate)
				}
			}
		}
	}
}
