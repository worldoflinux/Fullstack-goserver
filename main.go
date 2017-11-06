// Start program

package main
//importuri necesare
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/couchbase/gocb"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)
//Definire modele structura
type Movie struct {
	ID      string      `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Genre   string      `json:"genre,omitempty"`
	Formats MovieFormat `json:"formats,omitempty"`
}

type MovieFormat struct {
	Digital bool `json:"digital,omitempty"`
	Bluray  bool `json:"bluray,omitempty"`
	Dvd     bool `json:"dvd,omitempty"`
}

var bucket *gocb.Bucket
var bucketName string

//Definire endpointuri
func ListEndpoint(w http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	var movies []Movie
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "`")
	query.Consistency(gocb.RequestPlus)
	rows, _ := bucket.ExecuteN1qlQuery(query, nil)
	var row Movie
	for rows.Next(&row) {
		movies = append(movies, row)
		row = Movie{}
	}
	if movies == nil {
		movies = make([]Movie, 0)
	}
	json.NewEncoder(w).Encode(movies)
}

func SearchEndpoint(w http.ResponseWriter, req *http.Request) {
	var movies []Movie
	params := mux.Vars(req)
	var n1qlParams []interface{}
	n1qlParams = append(n1qlParams, strings.ToLower(params["title"]))
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "` WHERE LOWER(name) LIKE '%' || $1 || '%'")
	query.Consistency(gocb.RequestPlus)
	rows, _ := bucket.ExecuteN1qlQuery(query, n1qlParams)
	var row Movie
	for rows.Next(&row) {
		movies = append(movies, row)
		row = Movie{}
	}
	if movies == nil {
		movies = make([]Movie, 0)
	}
	json.NewEncoder(w).Encode(movies)
}

func CreateEndpoint(w http.ResponseWriter, req *http.Request) {
	var movie Movie
	_ = json.NewDecoder(req.Body).Decode(&movie)
	bucket.Insert(uuid.NewV4().String(), movie, 0)
	json.NewEncoder(w).Encode(movie)
}

func main() {
	fmt.Println("Starting server at http://10.0.1.15/fullstack-api")
	cluster, err  := gocb.Connect("couchbase://10.0.1.15")
	if err != nil {
		fmt.Println("ERRROR CONNECTING TO CLUSTER TRY AGAIN:", err)
	}
	bucketName = "fullstack-api"
	bucket, _ = cluster.OpenBucket(bucketName, "catalinpopescu")
	router := mux.NewRouter()
    handlers.AllowedOrigins([]string{"localhost:4200"})
	router.HandleFunc("/fullstack-api", ListEndpoint).Methods("GET")
	router.HandleFunc("/fullstack-api", CreateEndpoint).Methods("POST")
	router.HandleFunc("/search/{title}", SearchEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":8091", handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD"}), handlers.AllowedOrigins([]string{"localhost:4200"}))(router)))
}
