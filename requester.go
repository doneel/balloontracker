package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type reading struct {
	Time        string
	Latitude    float64
	Longitude   float64
	Altitude    float64
	DME         float64
	VOR         float64
	U           float64
	V           float64
	W           float64
	Pressure    float64
	Temperature float64
	RH          float64
}

var TIME_LAYOUT string = "2006010215"

type flightPath struct {
	Readings  []reading
	FinalLat  float64
	FinalLong float64
	Timestamp time.Time
}

type location struct {
	Latitude  float64
	Longitude float64
}

func beginRequesting(locs []location, alts []int, interval time.Duration, db *sql.DB, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		go requestAll(locs, alts, db)
		for timer := range time.Tick(interval) {
			timer = timer
			go requestAll(locs, alts, db)
		}
	}()
}

func requestAll(locs []location, alts []int, db *sql.DB) {
	for _, loc := range locs {
		for _, altitude := range alts {
			fp := requestFlightPath(loc, altitude)
			fmt.Printf("%f, %f, %d, %v\n", loc.Latitude, loc.Longitude, altitude, *fp)
		}
	}
}

func requestFlightPath(loc location, altitude int) *flightPath {
	timeStamp := getCurrentTimestamp()
	timeStr := timeStamp.Format(TIME_LAYOUT)
	requestString := "http://weather.uwyo.edu/cgi-bin/balloon_traj?TIME=" + timeStr + "&FCST=0&POINT=none&TOP=" + fmt.Sprint(altitude) + "&OUTPUT=list&LAT=" + fmt.Sprint(loc.Latitude) + "&LON=" + fmt.Sprint(loc.Longitude)
	resp, err := http.Get(requestString)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	dataCleaner, re_err := regexp.Compile("<pre>\\s([0-9][\\s|\\S]*)\\s<\\/PRE>")
	if err != nil {
		log.Fatal(re_err)
	}

	match := dataCleaner.FindSubmatch(body[:])
	if len(match) <= 1 {
		log.Fatal("Regexed data bad")
	}
	individuals := strings.Split(string(match[1]), "\n")

	return loadFlight(individuals, timeStamp)
}

func loadFlight(individuals []string, t_stamp time.Time) *flightPath {
	readingsArr := make([]reading, len(individuals))
	for i, line := range individuals {
		if len(line) != 0 {
			newReading, create_err := initReading(line)
			if create_err != nil {
				log.Fatal("Couldn't create a reading from read data!")
			}
			readingsArr[i] = *newReading
		}
	}
	flight := flightPath{Readings: readingsArr, FinalLat: readingsArr[len(readingsArr)-1].Latitude, FinalLong: readingsArr[len(readingsArr)-1].Longitude, Timestamp: t_stamp}
	return &flight
}

func initReading(line string) (*reading, error) {
	//fmt.Println("line is: " , line);
	fields := strings.Fields(line)
	numFields := make([]float64, 11)
	for pos, val := range fields {
		if pos != 0 {
			fVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Fatal("Couldn't convert read values to floats!")
			}
			numFields[pos-1] = fVal
		}
	}

	return &reading{fields[0], numFields[0], numFields[1], numFields[2], numFields[3], numFields[4], numFields[5], numFields[6], numFields[7], numFields[8], numFields[9], numFields[9]}, nil

}

func getCurrentTimestamp() time.Time {
	roundDuration2, _ := time.ParseDuration("0h")

	now := time.Now().Add(-roundDuration2)
	nowHours := now.Hour()
	nowHours -= (nowHours / 6) * 6
	roundDuration, err := time.ParseDuration(strconv.Itoa(nowHours) + "h")
	if err != nil {
		log.Fatal("Couldn't generate duration")
	}
	adjusted := now.Add(-roundDuration)
	return adjusted
}
