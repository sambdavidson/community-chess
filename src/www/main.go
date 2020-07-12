package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
)

var (
	port      = flag.Int("port", 3000, "port to serve community chess webserver")
	staticDir = flag.String("static_dir", "src/www/static", "directory of static content")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*staticDir)))
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(*staticDir, "index.html"))
	})

	/* DONE! */
	log.Printf("Starting HTTP Server on Port: 0.0.0.0:%d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port), nil))
}
