package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

var client *storage.Client
var bucket *storage.BucketHandle

var notFoundFile = ""
var defaultFile = "index.html"

func init() {
	c, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Error initializing storage client: %s\n", err)
	}
	client = c
	bucket = client.Bucket(os.Getenv("BGCS_BUCKET"))
	if f := os.Getenv("BGCS_NOT_FOUND_FILE"); f != "" {
		notFoundFile = f
	}
	if f := os.Getenv("BGCS_DEFAULT_FILE"); f != "" {
		defaultFile = f
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func modifyPath(path string) string {
	if len(path) == 0 || path[len(path)-1] == '/' {
		path += defaultFile
	} else if filepath.Ext(path) == "" {
		path += "/" + defaultFile
	}
	return path
}

func handler(w http.ResponseWriter, r *http.Request) {
	path := modifyPath(r.URL.Path[1:])
	log.Default().Printf("fetching file: %s\n", path)

	obj := bucket.Object(path)
	attrs, err := obj.Attrs(r.Context())
	if err != nil {
		if err == storage.ErrObjectNotExist {
			if notFoundFile == "" {
				log.Default().Printf("file %s does not exist. Returning 404.\n",
					path)
				http.NotFound(w, r)
				return
			}
			obj = bucket.Object(notFoundFile)
			log.Default().Printf("file %s does not exist. Fetching %s instead.\n",
				path, notFoundFile)

			ars, err := obj.Attrs(r.Context())
			if err != nil {
				// Fallback file not found. Return 500.
				http.Error(w, "fallback file not found", http.StatusInternalServerError)
				return
			}
			attrs = ars
		} else {
			http.Error(w, fmt.Sprintf("Failed to get object attributes: %v", err), http.StatusInternalServerError)
			return
		}
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
