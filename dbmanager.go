package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func openDb(filePath string) *sql.DB {
	fmt.Println(filePath)
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		log.Fatal(err)
	}
	createDb(db)

	/*if _, err := os.Stat(file); os.IsNotExist(err) {
		return createDb(filePath)
	}
	*/
	return db
}

func createDb(db *sql.DB) {
	createStatement := `
	create table if not exists Path (
	    pid int PRIMARY KEY,
	    time text,
	    startLat real,
	    startLon real,
	    endLat real,
	    endLon real,
	    ceiling integer,
	    UNIQUE (time, startLat, startLon, ceiling) ON CONFLICT REPLACE
	);

	create table if not exists Checkpoint (
	    cpid integer PRIMARY KEY,
	    pid integer,
	    time text,
	    latitude real,
	    longitude real,
	    altitude real,
	    temperature real,
	    pressure real,
	    dme real,
	    vor real,
	    u real,
	    v real,
	    w real,
	    rh real,
	    foreign key(pid) references Path(pid)
	);
	`
	//TODO: add trigger to prevent double requesting same route
	_, err := db.Exec(createStatement)
	if err != nil {
		log.Printf("Statement failed:\t%s\n", createStatement)
		log.Fatal(err)
	}
}

func writeFlight(fp *flightPath, db *sql.DB) (int64, error) {
	result, err := db.Exec("INSERT INTO Path(time, startLat, startLon, endLat, endLon, ceiling) VALUES (?, ?, ?, ?, ?, ?)", fp.Timestamp, fp.StartLat, fp.StartLon, fp.EndLat, fp.EndLon, fp.Ceiling)
	if err != nil {
		return -1, err
	}
	pid, err2 := result.LastInsertId()
	if err2 == nil {
		for _, cp := range fp.Readings {
			_, err := insertCheckpoint(cp, pid, db)
			if err != nil {
				fmt.Println("Failed to insert checkpoint: %v", cp)
				if err2 == nil {
					err2 = err
				}
			}
		}
	}

	return pid, err2
}

func insertCheckpoint(cp reading, pid int64, db *sql.DB) (int64, error) {
	result, err := db.Exec("INSERT INTO Checkpoint(pid, time, latitude, longitude, altitude, temperature, pressure, dme, vor, u, v, w, rh) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", pid, cp.Time, cp.Latitude, cp.Longitude, cp.Altitude, cp.Temperature, cp.Pressure, cp.DME, cp.VOR, cp.U, cp.V, cp.W, cp.RH)
	if err != nil {
		return -1, err
	}
	cpid, err2 := result.LastInsertId()
	return cpid, err2

}
