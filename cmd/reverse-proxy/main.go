package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/signal"

	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Hosts []struct {
		Path string
		Host string
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT variable is not set")
	}

	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}
	defer configFile.Close()
	bytes, _ := ioutil.ReadAll(configFile)
	var config Config
	json.Unmarshal(bytes, &config)

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          512,
		MaxIdleConnsPerHost:   512,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	r := http.NewServeMux()

	r.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("I'm OK!\n"))
	})

	for _, host := range config.Hosts {
		r.HandleFunc(host.Path, func(res http.ResponseWriter, req *http.Request) {
			log.Println(req.RemoteAddr, req.Method, req.URL.String())
			urlstr := host.Host
			url, err := url.Parse(urlstr)
			if err != nil {
				log.Fatal("Error parsing URLs ", err)
			}
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxy.Transport = tr
			proxy.ServeHTTP(res, req)
		})
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		fmt.Println()
		log.Println("SIGINT captured. Closing server.")
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Printf("HTTP server Shutdown with error %v\n", err)
		}
	}()

	log.Printf("Listening at :%s", port)
	log.Println(server.ListenAndServe())

}
