package main

import(
	"net/http"
	"html/template"
	"fmt"
	"encoding/json"
	"log"
	"bytes"
	"io"
	"flag"
	"os"
	"io/ioutil"
	"sync"
)

var flightPaths *[]*flightPath
var flightPathsJson string
var filePermissions os.FileMode = 0664
var port = flag.String("port", "80", "Specify which port to use as a webserver")
var webPageFile = flag.String("outFile", "", "Specify where to write the generated webpage. This overrides any webserver settings and cancels the server.")

type pageData struct {
	JsonFlightData	string
}

func main(){
	flag.Parse()

	var waitGroup sync.WaitGroup
	initDataManager(updateFlightData, &waitGroup)
	fmt.Println("Datamanger initialized")

	if len(*webPageFile)  <= 0{
		fmt.Println("Now listening for http requests on port", ":" + *port)
		http.HandleFunc("/", requestHandler)
		http.ListenAndServe(":" + *port, nil)
	}
	waitGroup.Wait()
}

func writeWebPageToFile(fullFilePath string){
	var fileBuf bytes.Buffer
	executeInterfaceTemplate(&fileBuf)
	ioutil.WriteFile(fullFilePath, fileBuf.Bytes(), filePermissions)
	fmt.Println("Wrote updated webpage to", fullFilePath)
}

func executeInterfaceTemplate(destination io.Writer){
	t, err := template.ParseFiles("interface.html")
	if err != nil{
		log.Fatal(err)
	}
	page := &pageData{JsonFlightData: flightPathsJson}
	err = t.Execute(destination, page)
}

func requestHandler(writer http.ResponseWriter, request *http.Request){
	executeInterfaceTemplate(writer)
}

func updateFlightData(updatedList *[]*flightPath){
	flightPaths = updatedList
	jsonArray, error := json.Marshal(flightPaths)
	if error != nil {
		log.Fatal(error)
	}
	var jsonBuffer  bytes.Buffer
	json.HTMLEscape(&jsonBuffer, jsonArray)

//	fmt.Println(string(jsonArray))
	flightPathsJson = jsonBuffer.String()//string(jsonArray)
	if len(*webPageFile)  > 0{
		writeWebPageToFile(*webPageFile)
	}
}
