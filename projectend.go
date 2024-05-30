package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const sisdataPath = "sisdata"
const basePath = "/api"

var Db *sql.DB

type Sisdata struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	LName     string `json:"lname"`
	Nostudent string `json:"nostudent"`
	Group     string `json:"group"`
	Branch    string `json:"branch"`
}

func SetupDB() {
	var err error
	Db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/projectendgo")

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(Db)
	Db.SetConnMaxLifetime(time.Minute * 3)
	Db.SetMaxOpenConns(10)
	Db.SetMaxIdleConns(10)
}

func getSisdataList() ([]Sisdata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	results, err := Db.QueryContext(ctx, `SELECT
	id,
	name,
	lname,
	nostudent,
	`+"`group`"+`,
	branch
	FROM sisdata`)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer results.Close()
	sisdatas := make([]Sisdata, 0)
	for results.Next() {
		var sisdata Sisdata
		results.Scan(&sisdata.ID,
			&sisdata.Name,
			&sisdata.LName,
			&sisdata.Nostudent,
			&sisdata.Group,
			&sisdata.Branch)

		sisdatas = append(sisdatas, sisdata)
	}
	return sisdatas, nil
}

func insertProduct(sisdata Sisdata) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := Db.ExecContext(ctx, `INSERT INTO sisdata
	(id,
	name,
	lname,
	nostudent,
	`+"`group`"+`,
	branch
	) VALUES (?, ?, ?, ?, ?, ?)`,
		sisdata.ID,
		sisdata.Name,
		sisdata.LName,
		sisdata.Nostudent,
		sisdata.Group,
		sisdata.Branch)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	insertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	return int(insertID), nil
}

func getSisdata(id int) (*Sisdata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := Db.QueryRowContext(ctx, `SELECT
	id,
	name,
	lname,
	nostudent,
	`+"`group`"+`,
	branch
	From sisdata
	WHERE id = ?`, id)

	sisdata := &Sisdata{}
	err := row.Scan(
		&sisdata.ID,
		&sisdata.Name,
		&sisdata.LName,
		&sisdata.Nostudent,
		&sisdata.Group,
		&sisdata.Branch)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Println(err)
		return nil, err
	}
	return sisdata, nil

}

func updateSisdata(sisdata Sisdata) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Db.ExecContext(ctx, `UPDATE sisdata SET
		name = ?,
		lname = ?,
		nostudent = ?,
		`+"`group`"+` = ?,
		branch = ?
	WHERE id = ?`,
		sisdata.Name,
		sisdata.LName,
		sisdata.Nostudent,
		sisdata.Group,
		sisdata.Branch,
		sisdata.ID)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func removeSisdata(ID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Db.ExecContext(ctx, `DELETE FROM sisdata where id = ?`, time.Duration(ID))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func handleSisdatas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sisdataList, err := getSisdataList()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		j, err := json.Marshal(sisdataList)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(j)
		if err != nil {
			log.Fatal(err)
		}

	case http.MethodPost:
		var sisdata Sisdata
		err := json.NewDecoder(r.Body).Decode(&sisdata)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		SisdataID, err := insertProduct(sisdata)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"id":%d}`, SisdataID)))

	case http.MethodOptions:
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleSisdata(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", sisdataPath))
	if len(urlPathSegments[1:]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	SisdataID, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		sisdata, err := getSisdata(SisdataID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if sisdata == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		j, err := json.Marshal(sisdata)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = w.Write(j)
		if err != nil {
			log.Fatal(err)
		}

	case http.MethodPut:
		var sisdata Sisdata
		err := json.NewDecoder(r.Body).Decode(&sisdata)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sisdata.ID = SisdataID
		err = updateSisdata(sisdata)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		err := removeSisdata(SisdataID)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)

	}

}

func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		handler.ServeHTTP(w, r)
	})
}

func SetupRoutes(apiBasePath string) {
	sisdataHandler := http.HandlerFunc(handleSisdata)
	http.Handle(fmt.Sprintf("%s/%s/", apiBasePath, sisdataPath), corsMiddleware(sisdataHandler))
	sisdatasHandler := http.HandlerFunc(handleSisdatas)
	http.Handle(fmt.Sprintf("%s/%s", apiBasePath, sisdataPath), corsMiddleware(sisdatasHandler))

}

func main() {

	SetupDB()
	SetupRoutes(basePath)
	log.Fatal(http.ListenAndServe(":5000", nil))

}
