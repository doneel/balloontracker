package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var pathExtension string = "/.balloontracker/saves/"
var layout string = "2006010215"

/*
 * TODO:  Timestamping the files
               -file name is time stamp
               -when saving, use timestamp as filename
               -store the file format as a constant
          Changing the dates in the regex
          File Saving
*/

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

type flightPath struct {
	Readings  []reading
	FinalLat  float64
	FinalLong float64
	Timestamp time.Time
}


/*func main() {
	flights := loadAllSaves(pathExtension)
	individuals, newTime := reqFlightPlan()
	timeStamp, err := time.Parse(layout, newTime)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(timeStamp.Format(layout))
	flights = append(flights, loadFlight(individuals, timeStamp))
	//fmt.Println(flight.Readings[0].toString())
	writeAllSaves(flights)

}*/

func initDataManager(callback func(*[]*flightPath), waitGroup *sync.WaitGroup) {
      flights := loadAllSaves(pathExtension)
      callback(&flights)
      duration, err := time.ParseDuration("6h")
      if err != nil{
          log.Fatal(err)
      }

      waitGroup.Add(1)
      go beginDataRequestInterval(duration, flights, callback, waitGroup)
}

func beginDataRequestInterval(duration time.Duration, flights []*flightPath, callback func(*[]*flightPath), waitGroup *sync.WaitGroup) {
      defer waitGroup.Done()
      fmt.Println("Timer initialized. Will request data.")
      makeDataRequest(flights, callback)
      for timer := range time.Tick(duration){
            timer = timer
            makeDataRequest(flights, callback)
        }
}

func makeDataRequest(flights []*flightPath, callback func(*[]*flightPath)){
        individuals, newTime := reqFlightPlan()
        timeStamp, err := time.Parse(layout, newTime)
        if err != nil {
            log.Fatal(err)
        }
        newFlight := loadFlight(individuals, timeStamp)
        writeAllSaves([]*flightPath{newFlight})
        flights = append(flights, newFlight)
        callback(&flights)
}

func getCurrentTimestamp() time.Time {
	//     roundDuration, err := time.ParseDuration("6h");
	roundDuration2, _ := time.ParseDuration("0h")

	now := time.Now().Add(-roundDuration2)
	nowHours := now.Hour()
	nowHours -= (nowHours / 6) * 6
	// fmt.Println(strconv.Itoa(nowHours) + "h")
	roundDuration, err := time.ParseDuration(strconv.Itoa(nowHours) + "h")
	if err != nil {
		log.Fatal("Couldn't generate duration")
	}
	//     fmt.Println("now ", now.Format(layout))
	//     fmt.Println(now.Truncate(roundDuration2).Truncate(roundDuration))
	adjusted := now.Add(-roundDuration)
	//fmt.Println(adjusted.Format(layout))
	return adjusted
	//squashedHours := (now.Hours()/6)*6
	//fmt.Println(squashedHours)
	//now.Hours = squashedHours
}

func loadAllSaves(path string) []*flightPath {
	layout := "2006010215"
	var saveList []*flightPath

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fileList, err := ioutil.ReadDir(usr.HomeDir + path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range fileList {
		fmt.Println("Loading file ", file.Name())
		timestamp, err := time.Parse(layout, file.Name())
		if err != nil {
			log.Fatal(err)
		}

		fileBytes, err := ioutil.ReadFile(usr.HomeDir + path + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(string(fileBytes))
		individuals := strings.Split(string(fileBytes), "\n")
		saveList = append(saveList, loadFlight(individuals, timestamp))
	}
	return saveList
}

func writeAllSaves(flightPaths []*flightPath) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	//writeDir, err := ioutil.ReadDir(usr.HomeDir + pathExtension)
	if err != nil {
		log.Fatal(err)
	}
	for _, flight := range flightPaths {
		timestampName := flight.Timestamp.Format(layout)
		fmt.Println("Saving flight ", timestampName)
		writePath := usr.HomeDir + pathExtension + timestampName
		var buf bytes.Buffer
		for _, line := range flight.Readings {
			buf.Write([]byte(line.toString()))
			buf.Write([]byte("\n"))
		}
		//fmt.Printf(writeDir[0].Name(), writePath)
		//fmt.Println(buf.String())
		var filePermissions os.FileMode = 0664
		ioutil.WriteFile(writePath, buf.Bytes(), filePermissions)
	}

}

func reqFlightPlan() ([]string, string) {
	timeStr := getCurrentTimestamp().Format(layout) //"2014011318"
	fmt.Println("Requesting from server flight path ", timeStr)

	resp, err := http.Get("http://weather.uwyo.edu/cgi-bin/balloon_traj?TIME=" + timeStr + "&FCST=0&POINT=none&TOP=30000&OUTPUT=list&LAT=45.220&LON=-111.761")

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

	return individuals, timeStr
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

func (read *reading) toString() string {
	var strBuffer bytes.Buffer

	typ := reflect.ValueOf(read).Elem()

	/*
	   typ := reflect.TypeOf(read)
	   if typ.Kind() == reflect.Ptr{
	         typ = typ.Elem()
	   }
	*/
	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i)
		fmt.Fprintf(&strBuffer, "%v ", p.Interface())
	}
	/*          fmt.Println(dType)
	    str += name + " "
	}
	*/

	return strBuffer.String()
}

func loadFlight(individuals []string, t_stamp time.Time) *flightPath {
	readingsArr := make([]reading, 0)
	for _, line := range individuals {
		if len(line) != 0 {
			newReading, create_err := initReading(line)
			if create_err != nil {
				log.Fatal("Couldn't create a reading from read data!")
			}
			readingsArr = append(readingsArr, *newReading)
		}
	}
	flight := flightPath{Readings: readingsArr, FinalLat: readingsArr[len(readingsArr)-1].Latitude, FinalLong: readingsArr[len(readingsArr)-1].Longitude, Timestamp: t_stamp}
	//fmt.Println(flight.Timestamp.Format(layout))
	return &flight
}

func splitIntoLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
