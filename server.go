package main

import (
	"bytes"
	"database/sql"
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

type jsonPath struct {
	Time     string
	StartLat float64
	StartLon float64
	EndLat   float64
	EndLon   float64
	Ceiling  float64
}

var flightPaths *[]*flightPath
var flightPathsJson string
var filePermissions os.FileMode = 0664
var port = flag.String("port", "8080", "Specify which port to use as a webserver")
var webPageFile = flag.String("outFile", "", "Specify where to write the generated webpage. This overrides any webserver settings and cancels the server.")

var START_LOCATIONS = []location{
	location{Latitude: 45.220, Longitude: -111.761},
}

var ALTITUDES = []int{8000, 10000, 12000, 15000, 18000, 20000, 30000, 50000}

var REQUEST_INTERVAL_STRING = "6h"

var DB_FILE string = "/.balloontrackerDb.db"

var MAIN_INTERFACE_PAGE = "html/interface.html"

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
	fmt.Printf("Datamanger initialized at %v\n", time.Now())

	http.HandleFunc("/", homepageHandler)
	http.HandleFunc("/paths/", func(writer http.ResponseWriter, request *http.Request) {
		pathsHandler(writer, request, db)
	})

	assetServer := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", assetServer))
	http.ListenAndServe(":"+*port, nil)

	/*	if len(*webPageFile) <= 0 {
			fmt.Println("Now listening for http requests on port", ":"+*port)
			http.HandleFunc("/", requestHandler)
			http.ListenAndServe(":"+*port, nil)
		}
	*/
	waitGroup.Wait()
}

func homepageHandler(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, MAIN_INTERFACE_PAGE)
}

func pathsHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	fpListP, err := getAllPaths(db)
	if err != nil {
		fmt.Printf("Error getting paths from db: %v\n", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonArr, err := json.Marshal(*fpListP)
	fmt.Printf("raw: %v\n\nmarshaled: %s\n\n\n\n", *fpListP, jsonArr)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(jsonArr)
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
