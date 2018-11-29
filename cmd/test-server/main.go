package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	port string
)

func main() {
	port = os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT variable is not set")
	}

	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()
	response, err := json.Marshal("ok")
	if err != nil {
		log.Println("Error marshal")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
	log.Println(port)
}
