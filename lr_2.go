package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func main() {
	if len(os.Args) > 1 {
		switch command := strings.ToLower(os.Args[1]); {

		case command == "--help":
			printHelp()

		case command == "--createdb":
			if _, err := os.Stat("./monitors.txt"); os.IsNotExist(err) {
				fmt.Println("ERROR! File \"monitors.txt\" does not exist!")
				return
			}

			if _, err := os.Stat("./products.db"); err == nil {
				err = os.Remove("./products.db")
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			CreateDB()
			AddMonitorsFromFile("./monitors.txt")

			fmt.Println("OK. File products.db is created!")
			return
		case command == "--start":
			http.HandleFunc("/category/monitors", GetMonitors)
			http.HandleFunc("/category/monitor/", GetStatForMonitor)
			http.HandleFunc("/category/monitor_click/", AddClickForMonitor)
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "LR_2.html")
			})
			fmt.Println("The server is running!")
			fmt.Println("Looking forward to request...")
			if err := http.ListenAndServe("localhost:8030", nil); err != nil {
				log.Fatal("Failed to start server!", err)
			}
		default:
			{
				printHelp()
			}
		}
	}
}

func printHelp() {
	fmt.Println()
	fmt.Println("Help: 						./counter --help")
	fmt.Println("Create products database:	./counter --createdb")
	fmt.Println("Start server:				./counter --start")
	fmt.Println()
}

func CreateDB() {
	OpenDB()
	_, err := DB.Exec("create table monitors(id integer, name varchar(255) not null, count integer)")
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	DB.Close()
}

func OpenDB() {
	db, err := sql.Open("sqlite3", "products.db")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	DB = db
}

func AddMonitorsFromFile(filename string) {
	var file *os.File
	var err error
	if file, err = os.Open(filename); err != nil {
		log.Fatal("Failed to open the file: ", err)
		os.Exit(2)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	OpenDB()

	for scanner.Scan() {
		arr := strings.Split(scanner.Text(), ",")
		id := arr[0]
		monitorName := arr[1]
		_, err = DB.Exec("insert into monitors(id, name, count) values($1, $2, 0)", id, monitorName)
	}
}

func AddClickForMonitor(w http.ResponseWriter, request *http.Request) {

	err := request.ParseForm()

	if err != nil {
		fmt.Fprintf(w, "{%s}", err)
	} else {
		monitorId := strings.TrimPrefix(request.URL.Path, "/category/monitor_click/")
		OpenDB()
		countValue := 0
		rows, _ := DB.Query("select count from monitors where id=" + monitorId)
		for rows.Next() {
			rows.Scan(&countValue)
		}
		countValue++
		_, err = DB.Exec("update monitors set count=" + strconv.Itoa(countValue) + " where id=" + monitorId)
	}
}

func GetFromDBNameModel(tblName string) []string {
	var arr []string
	var monitorName string
	rows, _ := DB.Query("select name from " + tblName)
	for rows.Next() {
		rows.Scan(&monitorName)
		arr = append(arr, monitorName)
	}
	return arr
}

func GetMonitors(w http.ResponseWriter, request *http.Request) {
	OpenDB()
	monitors := GetFromDBNameModel("monitors")
	err := request.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "{%s}", err)
	} else {
		var monitorList [][]interface{}
		for i := range monitors {
			id := i + 1 // Визначення ID монітора
			name := monitors[i]
			monitorList = append(monitorList, []interface{}{id, name})
		}

		responseJSON := map[string]interface{}{
			"monitors": monitorList,
		}

		jsonData, err := json.Marshal(responseJSON)
		if err != nil {
			fmt.Fprintf(w, `{ "error": "Error processing JSON" }`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func GetStatForMonitor(w http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "{%s}", err)
	} else {
		countValue := 0
		monitorId := strings.TrimPrefix(request.URL.Path, "/category/monitor/")
		OpenDB()
		rows, _ := DB.Query("select count from monitors where id=" + monitorId)
		for rows.Next() {
			rows.Scan(&countValue)
		}
		strOut := "{ \"id\": \"" + monitorId + "\", \"count\":\"" + strconv.Itoa(countValue) + "\"}"
		fmt.Fprintf(w, strOut)
	}
}
