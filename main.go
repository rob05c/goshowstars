package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"time"
//	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"html/template"
)

const version = "0.0.0"

const indexFile = "index.html"
const starTemplateFile = "startemplate.html"

var port uint
var dataservice string
var filePath string

func init() {
	const (
		portDefault          = 0
		portUsage            = "http serve port"
		dataserviceDefault   = ""
		dataserviceUsage     = "URL for data service"
		filePathDefault      = ""
		filePathUsage        = "Working directory, for html files"
	)
	flag.UintVar(&port, "port", portDefault, portUsage)
	flag.UintVar(&port, "p", portDefault, portUsage+" (shorthand)")
	flag.StringVar(&dataservice, "data-service", dataserviceDefault, dataserviceUsage)
	flag.StringVar(&dataservice, "d", dataserviceDefault, dataserviceUsage+" (shorthand)")
	flag.StringVar(&filePath, "files", filePathDefault, filePathUsage)
	flag.StringVar(&filePath, "f", filePathDefault, filePathUsage+" (shorthand)")
}

func printUsage() {
	exeName := os.Args[0]
	fmt.Println(exeName + " " + version + " usage: ")
	fmt.Println("\t" + exeName + "-d data-service-url -p serve-port -f html-files-path")
	fmt.Println("flags:")
	flag.PrintDefaults()
	fmt.Println("example:\n\t" + exeName + "-d 192.168.0.42 -p 80 -f data")
}

type Star struct {
	Id                int64   `json:"id"`
	Name              string  `json:"name"`
	X                 float64 `json:"x"`
	Y                 float64 `json:"y"`
	Z                 float64 `json:"z"`
	Color             float32 `json:"color"`
	AbsoluteMagnitude float32 `json:"absolute-magnitude"`
	Spectrum          string  `json:"spectrum"`
}

func (star *Star) Json() []byte {
	bytes, err := json.Marshal(star)
	if err != nil {
		return nil ///< @todo fix to return JSON error
	}
	return bytes
}

func JsonToStar(starjson []byte) (Star, error) {
	var star Star
	err := json.Unmarshal(starjson, &star)
	if err != nil {
		return Star {}, err ///< @todo fix to return JSON error
	}
	return star, nil
}


type NullStar struct {
	Id                sql.NullInt64
	Name              sql.NullString
	X                 sql.NullFloat64
	Y                 sql.NullFloat64
	Z                 sql.NullFloat64
	Color             sql.NullFloat64
	AbsoluteMagnitude sql.NullFloat64
	Spectrum          sql.NullString
}

/// returns a Star, with zero values for any null values in the NullStar
func (nstar *NullStar) Star() Star {
	var star Star
	if nstar.Id.Valid {
		star.Id = nstar.Id.Int64
	}
	if nstar.Name.Valid {
		star.Name = nstar.Name.String
	}
	if nstar.X.Valid {
		star.X = nstar.X.Float64
	}
	if nstar.Y.Valid {
		star.Y = nstar.Y.Float64
	}
	if nstar.Z.Valid {
		star.Z = nstar.Z.Float64
	}
	if nstar.Color.Valid {
		star.Color = float32(nstar.Color.Float64)
	}
	if nstar.AbsoluteMagnitude.Valid {
		star.AbsoluteMagnitude = float32(nstar.AbsoluteMagnitude.Float64)
	}
	if nstar.Spectrum.Valid {
		star.Spectrum = nstar.Spectrum.String
	}
	return star
}

func getstar(id int64) (Star, error) {
	response, err := http.Get("http://" + dataservice + "/star/" + strconv.FormatInt(id, 10))
	if err != nil {
		return Star{}, err
	}

	starjson, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return Star{}, err
	}

	star, err := JsonToStar(starjson)
	if err != nil {
		return Star{}, err
	}

	return star, nil
}

func reachable(uri string) bool {
	return true
	timeout := time.Second * 10
	_, err := net.DialTimeout("tcp", uri, timeout)
	return err == nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if dataservice == "" {
		printUsage()
		return
	}

	if(len(dataservice) > len("http://") && dataservice[0:len("http://")] == "http://") {
		dataservice = dataservice[len("http://"):]
	}

	if !reachable(dataservice) {
		fmt.Println("Could not reach data service " + dataservice)
		return
	}

	indexHtml, err := ioutil.ReadFile(filePath + "/" + indexFile)
	if err != nil {
		fmt.Println("Error reading index file: ")
		fmt.Println(err)
		return
	}

	starTemplate, err := template.ParseFiles(filePath + "/" + starTemplateFile)
	if err != nil {
		fmt.Println("Error reading star template file: ")
		fmt.Println(err)
		return
	}

	serveIndex := func(w http.ResponseWriter) error {
		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Content-Length", strconv.Itoa(len(indexHtml)))
		_, err = w.Write(indexHtml)
		return err
	};

	serveStar := func(w http.ResponseWriter, id int64) error {
		star, err := getstar(id)
		if err != nil {
			return err
		}
		return starTemplate.Execute(w, star)
	};


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Serving request to " + r.URL.Path)
		err := serveIndex(w)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	http.HandleFunc("/star/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Serving request to " + r.URL.Path)

		staridStr := r.URL.Path[len("/star/"):]

		starid, err := strconv.ParseInt(staridStr, 10, 64)
		_, err = strconv.Atoi(staridStr)
		if err != nil {
			serveIndex(w)
		}

		err = serveStar(w, starid)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	fmt.Println("Serving on " + strconv.Itoa(int(port)) + "...")
	http.ListenAndServe(":"+strconv.Itoa(int(port)), nil)
}
