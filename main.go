package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
)

var client *storage.Client
var bucket *storage.BucketHandle

var notFoundFile = ""
var defaultFile = ""

func init() {
	c, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Error initializing storage client: %s\n", err)
	}
	client = c
	bucket = client.Bucket(os.Getenv("BGCS_BUCKET"))
	notFoundFile = os.Getenv("BGCS_NOT_FOUND_FILE")
	if defaultFile = os.Getenv("BGCS_DEFAULT_FILE"); defaultFile == "" {
		defaultFile = "index.html"
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if len(path) == 0 || path[len(path)-1] == '/' {
		path += "index.html"
	}

	obj := bucket.Object(path)
	attrs, err := obj.Attrs(r.Context())
	if err != nil {
		if err == storage.ErrObjectNotExist {
			if notFoundFile == "" {
				http.NotFound(w, r)
			}
			obj = bucket.Object(notFoundFile)
			//lint:ignore SA4006 Not unused
			attrs, _ = obj.Attrs(r.Context())
		} else {
			http.Error(w, fmt.Sprintf("Failed to get object attributes: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", attrs.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", attrs.Size))

	rc, err := obj.NewReader(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read object: %v", err), http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	if _, err := rc.WriteTo(w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write object to response: %v", err), http.StatusInternalServerError)
		return
	}
}
