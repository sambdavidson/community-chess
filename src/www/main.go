package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 3000, "port to serve community chess webserver.")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	// Obviously terribly insecure. Fix later
	http.Handle("/", http.FileServer(http.Dir(".")))
	/* DONE! */
	log.Printf("Starting HTTP Server on Port: 0.0.0.0:%d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port), nil))
}
