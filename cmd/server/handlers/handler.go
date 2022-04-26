package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	// extract metric from url
	url := r.URL.Path
	splittedUrl := strings.Split(url, "/")
	if splittedUrl[1] == "update" {
		//процессим дальше
		if splittedUrl[2] == "gauge" || splittedUrl[2] == "counter" {
			fmt.Println(splittedUrl[2], splittedUrl[3], splittedUrl[4])
		}
	}

	// write metric to repository

	// response answer
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	// намеренно сделана ошибка в JSON
	w.Write([]byte(`Metric updated`))
}
