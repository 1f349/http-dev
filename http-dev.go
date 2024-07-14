package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var listenFlag, targetFlag, hostFlag, allowCors string

func main() {
	flag.StringVar(&listenFlag, "listen", ":8080", "Listen address")
	flag.StringVar(&targetFlag, "target", "", "Set target URL to forward to")
	flag.StringVar(&hostFlag, "host", "", "Set the host value")
	flag.StringVar(&allowCors, "cors", "", "Set cross origin domain or '*' for all origins")
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
			r.Out.Header = r.In.Header.Clone()
		},
		ModifyResponse: func(resp *http.Response) error {
			if allowCors != "" {
				resp.Header.Set("Access-Control-Allow-Origin", allowCors)
				resp.Header.Set("Access-Control-Allow-Credentials", "true")
				resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
				resp.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, HEAD, DELETE")

				if resp.Request.Method == http.MethodOptions {
					resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
					resp.Header.Set("X-Content-Type-Options", "nosniff")
					resp.StatusCode = http.StatusNoContent
					resp.Status = http.StatusText(http.StatusNoContent)
				}
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
	if allowCors != "" {
		log.Println("- Allowing all cross-origin requests")
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
