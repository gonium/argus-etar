package main

import (
	"bufio"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/gonium/argus_etar"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var argus_flights argus.Flights
var TWITTER_CONSUMER_KEY = os.Getenv("CONSUMER_KEY")
var TWITTER_CONSUMER_SECRET = os.Getenv("CONSUMER_SECRET")
var TWITTER_ACCESS_TOKEN = os.Getenv("ACCESS_TOKEN")
var TWITTER_ACCESS_TOKEN_SECRET = os.Getenv("ACCESS_TOKEN_SECRET")
var api *anaconda.TwitterApi

func receive_SBS(netaddress string, waitgroup sync.WaitGroup) {
	defer waitgroup.Done()
	conn, err := net.Dial("tcp", netaddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot connect to SBS stream:", err.Error())
		os.Exit(1)
	}
	connbuf := bufio.NewReader(conn)
	fmt.Println("Collecting incoming flight facts.")
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
				recvTime := time.Now()
				hexIdent := sbs[4]
				switch sbs[1] { // This contains the subtype
				case "1": // ES Identification
					//					fmt.Println("Hexident:", hexIdent, "Callsign:", sbs[10])
					argus_flights.AddCallsign(hexIdent, recvTime, sbs[10])
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
					argus_flights.AddPosition(hexIdent, recvTime, alt, lat, lon)
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
					argus_flights.AddVelocity(hexIdent, recvTime, groundspeed,
						verticalrate)
				}
			}
		}
	}
}

func evaluateFlights(waitgroup sync.WaitGroup, quit chan struct{}) {
	defer waitgroup.Done()
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			// to stuff
			fmt.Println("Current flights")
			fmt.Println(argus_flights)
			for _, value := range argus_flights {
				if value.IsETAR() {
					tweet := "Incoming: Flight with unknown callsign."
					if len(strings.TrimSpace(value.Callsign)) != 0 {
						tweet = fmt.Sprintf(
							"Incoming: Flight %s, http://flightaware.com/live/flight/%s",
							value.Callsign, value.Callsign)
					}
					tweetID, err := api.PostTweet("I'm alive!", nil)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to post tweet: %s\n", err.Error())
					} else {
						fmt.Println(tweetID, ":", tweet)
					}
				}
			}
			argus_flights.Tick()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func init() {
	// Initialize our flight surveillance recorder data structure
	argus_flights = make(map[string]argus.Flight)
	if TWITTER_CONSUMER_KEY == "" || TWITTER_CONSUMER_SECRET == "" ||
		TWITTER_ACCESS_TOKEN == "" || TWITTER_ACCESS_TOKEN_SECRET == "" {
		fmt.Fprintf(os.Stderr, "Credentials are invalid: at least one is empty\n")
		fmt.Fprintf(os.Stderr, "use export TWITTER_ACCESS_TOKEN=<Your token> etc.\n")
		os.Exit(1)
	}
	anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
	api = anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_TOKEN_SECRET)
}

func main() {
	// TODO: Terminate goroutines properly (CTRL-C, Errors)
	wg := sync.WaitGroup{}
	wg.Add(2)
	quit := make(chan struct{})
	// TODO: Take this from the command line
	go receive_SBS("127.0.0.1:30003", wg)
	go evaluateFlights(wg, quit)

	//}

	wg.Wait()
}
