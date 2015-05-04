package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"sync"
	"time"
)

var flightPaths *[]*flightPath
var flightPathsJson string
var filePermissions os.FileMode = 0664
var port = flag.String("port", "80", "Specify which port to use as a webserver")
var webPageFile = flag.String("outFile", "", "Specify where to write the generated webpage. This overrides any webserver settings and cancels the server.")

var START_LOCATIONS = []location{
	location{Latitude: 45.220, Longitude: -111.761},
}

var ALTITUDES = []int{8000, 10000, 12000, 15000, 18000, 20000, 30000, 50000}

var REQUEST_INTERVAL_STRING = "6h"

var DB_FILE string = "/.balloontrackerDb.db"

type pageData struct {
	JsonFlightData string
}

func main() {
	flag.Parse()

	usr, _ := user.Current()
	dbPath := usr.HomeDir + DB_FILE
	db := openDb(dbPath)

	var waitGroup sync.WaitGroup
	intervalDuration, err := time.ParseDuration(REQUEST_INTERVAL_STRING)
	if err != nil {
		log.Fatal(err)
	}
	beginRequesting(START_LOCATIONS, ALTITUDES, intervalDuration, db, &waitGroup)
	//	initDataManager(updateFlightData, &waitGroup)
	fmt.Println("Datamanger initialized")

	/*	if len(*webPageFile) <= 0 {
			fmt.Println("Now listening for http requests on port", ":"+*port)
			http.HandleFunc("/", requestHandler)
			http.ListenAndServe(":"+*port, nil)
		}
	*/
	waitGroup.Wait()
}

func writeWebPageToFile(fullFilePath string) {
	var fileBuf bytes.Buffer
	executeInterfaceTemplate(&fileBuf)
	ioutil.WriteFile(fullFilePath, fileBuf.Bytes(), filePermissions)
	fmt.Println("Wrote updated webpage to", fullFilePath)
}

func executeInterfaceTemplate(destination io.Writer) {
	t, err := template.ParseFiles("interface.html")
	if err != nil {
		log.Fatal(err)
	}
	page := &pageData{JsonFlightData: flightPathsJson}
	err = t.Execute(destination, page)
}

func requestHandler(writer http.ResponseWriter, request *http.Request) {
	executeInterfaceTemplate(writer)
}

func updateFlightData(updatedList *[]*flightPath) {
	flightPaths = updatedList
	jsonArray, error := json.Marshal(flightPaths)
	if error != nil {
		log.Fatal(error)
	}
	var jsonBuffer bytes.Buffer
	json.HTMLEscape(&jsonBuffer, jsonArray)

	//	fmt.Println(string(jsonArray))
	flightPathsJson = jsonBuffer.String() //string(jsonArray)
	if len(*webPageFile) > 0 {
		writeWebPageToFile(*webPageFile)
	}
}
