package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var listenFlag, targetFlag, hostFlag string
var allowCors bool

func main() {
	flag.StringVar(&listenFlag, "listen", ":8080", "Listen address")
	flag.StringVar(&targetFlag, "target", "", "Set target URL to forward to")
	flag.StringVar(&hostFlag, "host", "", "Set the host value")
	flag.BoolVar(&allowCors, "cors", false, "Allow all cross origin requests")
	flag.Parse()

	target, err := url.Parse(targetFlag)
	if err != nil {
		log.Fatal("Failed to parse target URL:", err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			if hostFlag != "" {
				r.Out.Host = hostFlag
			}
		},
		ModifyResponse: func(response *http.Response) error {
			if !allowCors {
				return nil
			}

			response.Header.Add("Access-Control-Allow-Origin", "*")
			response.Header.Add("Access-Control-Allow-Credentials", "true")
			response.Header.Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

			if response.Request.Method == "OPTIONS" {
				_ = response.Body.Close()
				response.Header.Set("Content-Type", "text/plain; charset=utf-8")
				response.Header.Set("X-Content-Type-Options", "nosniff")
				response.StatusCode = http.StatusNoContent
				response.Status = http.StatusText(http.StatusNoContent)
				response.Body = io.NopCloser(strings.NewReader("No Content\n"))
			}
			return nil
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

	log.Printf("HTTP Dev listening on %s\n", listenFlag)
	log.Printf("- Forwarding to '%s'\n", targetFlag)
	if hostFlag != "" {
		log.Printf("- Rewriting host to '%s'\n", hostFlag)
	}
	if allowCors {
		log.Println("- Allowing all cross-origin requests")
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
