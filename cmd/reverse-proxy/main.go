package main

import (
	"log"
	"net"

	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	_nHosts int
	_port   int
	count   int
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT variable is not set")
	}
	var err error
	_port, err = strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	nHosts := os.Getenv("HOSTS")
	if nHosts == "" {
		log.Fatal("HOSTS variable is not set")
	}
	_nHosts, err = strconv.Atoi(nHosts)
	if err != nil {
		log.Fatal(err)
	}
	count = 0

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
	pxs := initProxy(tr)

	http.HandleFunc("/api/", func(res http.ResponseWriter, req *http.Request) {
		count = (count + 1) % _nHosts
		pxs[count].ServeHTTP(res, req)
	})

	if err = http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func initProxy(tr *http.Transport) []*httputil.ReverseProxy {
	prxs := []*httputil.ReverseProxy{}
	for i := 1; i <= _nHosts; i++ {
		p := (_port + i) % 65536
		if p < 1024 {
			p += 1024
		}
		urlstr := "http://localhost" + ":" + strconv.Itoa(p)
		url, err := url.Parse(urlstr)
		if err != nil {
			log.Fatal("Error parsing URLs ", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.Transport = tr
		prxs = append(prxs, proxy)
	}
	return prxs
}
