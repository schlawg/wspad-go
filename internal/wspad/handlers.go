package wspad

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	indexTemplate *template.Template
)

func Init() {
	var err error
	indexTemplate, err = template.ParseFiles(filepath.Join(TemplatesDir, "index.html"))
	if err != nil {
		log.Fatalf("Failed to parse index template: %v", err)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files, err := os.ReadDir(DataDir)
	if err != nil {
		log.Printf("Error reading data directory: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pads []string
	for _, file := range files {
		if !file.IsDir() {
			pads = append(pads, file.Name())
		}
	}

	data := struct {
		Pads []string
	}{
		Pads: pads,
	}

	if err := indexTemplate.Execute(w, data); err != nil {
		log.Printf("Error executing index template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func PadHandler(w http.ResponseWriter, r *http.Request) {
	padName := strings.TrimPrefix(r.URL.Path, "/")

	if padName == "" || strings.Contains(padName, "/") || strings.Contains(padName, "..") {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filepath.Join(StaticDir, "board.html"))
}
