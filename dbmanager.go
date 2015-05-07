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
	    pid integer PRIMARY KEY,
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

func getAllPaths(db *sql.DB) (*[]jsonPath, error) {
	paths := make([]jsonPath, 0)
	rows, err := db.Query(`SELECT * FROM Path GROUP BY ceiling ORDER BY time DESC`)
	if err != nil {
		return &paths, err
	}
	defer rows.Close()

	for rows.Next() {
		var fp jsonPath
		var id int64

		if err := rows.Scan(&id, &fp.Time, &fp.StartLat, &fp.StartLon, &fp.EndLat, &fp.EndLon, &fp.Ceiling); err != nil {
			return &paths, err
		}
		cpsP, err := getCheckpoints(db, id)
		if err != nil {
			fmt.Printf("Error parsing checkpoint: %v: %v\n", err, *cpsP)
		}
		fp.Checkpoints = *cpsP
		paths = append(paths, fp)

	}
	if err := rows.Err(); err != nil {
		return &paths, err
	}
	return &paths, nil
}

func getCheckpoints(db *sql.DB, pid int64) (*[]jsonCheckpoint, error) {
	points := make([]jsonCheckpoint, 0)
	rows, err := db.Query(`SELECT time, latitude, longitude, altitude, temperature FROM Checkpoint WHERE pid = ?`, pid)

	if err != nil {
		return &points, err
	}
	defer rows.Close()

	for rows.Next() {
		var cp jsonCheckpoint
		if err := rows.Scan(&cp.Time, &cp.Lat, &cp.Lon, &cp.Altitude, &cp.Temperature); err != nil {
			return &points, err
		}
		points = append(points, cp)
	}
	if err := rows.Err(); err != nil {
		return &points, err
	}
	return &points, nil
}
