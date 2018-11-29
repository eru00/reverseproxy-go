package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	c = 12
	N = 25000
	n = N / c
)

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1024,
			MaxIdleConnsPerHost: 1024,
		},
	}

	cFlag := flag.Int("c", 1, "concurrency")
	nFlag := flag.Int("n", 2500, "transactions")
	flag.Parse()
	c = *cFlag
	N = *nFlag
	fmt.Println(c)
	fmt.Println(N)
	url := "http://localhost:3000/api/"

	stop := make(chan bool, c)
	t0 := time.Now()
	for ci := 0; ci < c; ci++ {
		go func(cii int) {

			for ni := 0; ni < n; ni++ {
				curl1 := fmt.Sprintf("testestest")

				body := strings.NewReader(curl1)
				singleTest(cii, ni, client, url, body)
			}
			stop <- true
		}(ci)
	}
	for ci := 0; ci < c; ci++ {
		<-stop
		fmt.Print(".")
	}
	totTime := time.Now().Sub(t0).Seconds()
	fmt.Printf("\nTest time: %f s\n", totTime)
	fmt.Printf("TPS: %.1f\n", float64(c*n)/totTime)
}

func singleTest(c, n int, client *http.Client, url string, body io.Reader) {
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}
	var asd string
	re, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	err = json.Unmarshal(re, &asd)
	if err != nil {
		log.Println(err)
		return
	}
	// fmt.Println(asd)

	io.Copy(ioutil.Discard, resp.Body)

	return
}
