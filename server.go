package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

type application struct {
	outputsDir string
	nextID     atomic.Uint64
}

type reportRequest struct {
	Name string `json:"name"`
}

type reportResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	DownloadURL string `json:"downloadUrl"`
}

func newHandler(dataDir string) (http.Handler, error) {
	outputsDir := filepath.Join(dataDir, "outputs")
	if err := os.MkdirAll(outputsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	app := &application{outputsDir: outputsDir}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.health)
	mux.HandleFunc("POST /api/reports", app.createReport)
	mux.HandleFunc("GET /downloads/{filename}", app.downloadReport)

	return withCORS(mux), nil
}

func (app *application) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *application) createReport(w http.ResponseWriter, r *http.Request) {
	var request reportRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" || len(request.Name) > 80 {
		writeError(w, http.StatusBadRequest, "name must contain 1 to 80 characters")
		return
	}

	id := fmt.Sprintf("report-%d-%d", time.Now().UnixMilli(), app.nextID.Add(1))
	filename := id + ".csv"
	if err := writeCSV(filepath.Join(app.outputsDir, filename), request.Name); err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate report")
		return
	}

	writeJSON(w, http.StatusCreated, reportResponse{
		ID:          id,
		Name:        request.Name,
		Status:      "ready",
		DownloadURL: "/downloads/" + filename,
	})
}

func (app *application) downloadReport(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" || filepath.Base(filename) != filename || !strings.HasSuffix(filename, ".csv") {
		writeError(w, http.StatusBadRequest, "invalid report filename")
		return
	}

	path := filepath.Join(app.outputsDir, filename)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "report not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not read report")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, filename))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	http.ServeFile(w, r, path)
}

func writeCSV(path, reportName string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"report", "total"}); err != nil {
		return err
	}
	if err := writer.Write([]string{reportName, "42"}); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
