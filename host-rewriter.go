package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var listenFlag, targetFlag, hostFlag string

func main() {
	flag.StringVar(&listenFlag, "listen", ":8080", "Listen address")
	flag.StringVar(&targetFlag, "target", "", "Set target URL to forward to")
	flag.StringVar(&hostFlag, "host", "", "Set the host value")
	flag.Parse()

	target, err := url.Parse(targetFlag)
	if err != nil {
		log.Fatal("Failed to parse target URL:", err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Host = hostFlag
		},
	}
	s := http.Server{
		Addr:              listenFlag,
		Handler:           proxy,
		ReadTimeout:       150 * time.Second,
		ReadHeaderTimeout: 150 * time.Second,
		WriteTimeout:      150 * time.Second,
		IdleTimeout:       150 * time.Second,
		MaxHeaderBytes:    4096000,
	}

	log.Printf("Host Rewriter listening on %s\n", listenFlag)
	log.Printf("Rewriting host to '%s' and forwarding to '%s'\n", hostFlag, targetFlag)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
