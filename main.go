package main

import (
	"bytes"
	"errors"
	"github.com/yuin/goldmark"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func main() {
	http.HandleFunc("/articles/", ArticleHandler(NewMarkdownReader()))

	log.Println("Starting server on :3030")
	if err := http.ListenAndServe(":3030", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type MarkdownReader interface {
	FetchContent(slug string) (string, error)
}

type LocalFileReader struct {
	basePath string
}

func NewMarkdownReader() *LocalFileReader {
	return &LocalFileReader{basePath: "./content"}
}

func (r *LocalFileReader) FetchContent(slug string) (string, error) {
	safeSlug := sanitizeSlug(slug)
	filePath := filepath.Join(r.basePath, safeSlug+".md")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", errors.New("content not found")
	}
	return string(content), nil
}

func sanitizeSlug(slug string) string {
	return strings.Trim(filepath.Clean(slug), "/")
}

func ArticleHandler(reader MarkdownReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/articles/")
		if slug == "" {
			http.Error(w, "Article not specified", http.StatusBadRequest)
			return
		}

		markdownContent, err := reader.FetchContent(slug)
		if err != nil {
			http.Error(w, "Article not found", http.StatusNotFound)
			return
		}

		htmlContent, err := renderMarkdownToHTML(markdownContent)
		if err != nil {
			http.Error(w, "Error rendering article", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(htmlContent)
	}
}

func renderMarkdownToHTML(markdown string) ([]byte, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}