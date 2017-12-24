package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/rickbau5/anomaly-tracker-server/cmd/internal/tracker"
)

type response struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`

	StatusCode int `json:"-"`
}

var debugMode = false

func Init(debug bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthcheck)
	mux.HandleFunc("/anomaly", Authenticate(handleAnomaly))

	debugMode = debug

	return mux
}

func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := getAPIKey(r)
		if key == nil {
			writeErrorResponse("Invalid API key", http.StatusUnauthorized, w)
			return
		}
		// TODO: create a type alias for this
		ctx := context.WithValue(r.Context(), "api_key", *key)
		next(w, r.WithContext(ctx))
	}
}

func PathLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func handleAnomaly(w http.ResponseWriter, r *http.Request) {
	i := r.Context().Value("api_key")
	apiKey, ok := i.(tracker.APIKey)
	if !ok {
		log.Println("Got a weird api key")
		writeErrorResponse("Interanl error", http.StatusInternalServerError, w)
		return
	}

	var anomaly tracker.Anomaly

	if r.Method != http.MethodGet {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse("Failed reading body", http.StatusNotAcceptable, w)
			return
		}

		err = json.Unmarshal(bytes, &anomaly)
		if err != nil {
			writeErrorResponse("Failed reading anomaly", http.StatusNotAcceptable, w)
			return
		}
	}

	var resp response
	switch r.Method {
	case http.MethodPost:
		resp = addAnomaly(anomaly, apiKey)
	case http.MethodDelete:
		resp = deleteAnomaly(anomaly, apiKey)
	default:
		resp = response{
			Error:      "Unrecognized method: " + r.Method,
			StatusCode: http.StatusMethodNotAllowed,
		}
	}
	writeFullResponse(resp, w)
}

func addAnomaly(anomaly tracker.Anomaly, apiKey tracker.APIKey) response {
	err := tracker.AddAnomaly(anomaly, apiKey)
	if err != nil {
		log.Println("Failed adding anomaly:", err.Error())
		return response{
			Error:      sanitizeError(err),
			StatusCode: http.StatusNotAcceptable,
		}
	}

	return response{
		StatusCode: http.StatusCreated,
		Message:    "created",
	}
}

func deleteAnomaly(anomaly tracker.Anomaly, apiKey tracker.APIKey) response {
	err := tracker.DeleteAnomaly(anomaly, apiKey)
	if err != nil {
		log.Println("Failed deleting anomaly:", err.Error())
		return response{
			Error:      sanitizeError(err),
			StatusCode: http.StatusNotAcceptable,
		}
	}
	return response{
		StatusCode: http.StatusOK,
		Message:    "deleted",
	}
}

func writeErrorResponse(message string, status int, w http.ResponseWriter) {
	res := response{}
	res.Error = message
	w.WriteHeader(status)
	writeResponse(res, w)

	log.Println("Wrote error response:", message)
}

func writeResponse(res response, w http.ResponseWriter) {
	bytes, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		log.Println("Failed marshalling response:", err.Error())
		return
	}

	w.Write(bytes)
}

func writeFullResponse(res response, w http.ResponseWriter) {
	bytes, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	w.Write(bytes)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("^.^"))
}

func sanitizeError(err error) string {
	if debugMode {
		return err.Error()
	}
	if tracker.IsErrAnomaly(err) {
		return tracker.GetErrAnomalyMessage(err)
	}
	return "Internal error"
}

func getAPIKey(request *http.Request) *tracker.APIKey {
	apiKey := request.Header.Get("Authentication-Key")
	if apiKey == "" {
		log.Println("No api key specified.")
		return nil
	}

	key, err := tracker.CheckAPIKey(apiKey)
	if err != nil {
		log.Println("Unexpected error looking up key:", err.Error())
		return nil
	}
	if key == nil {
		log.Printf("Nothing found for key: '%s'\n", apiKey)
	}
	return key
}

func Method(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Method) != strings.ToLower(method) {
			writeErrorResponse(
				fmt.Sprintf("Wrong method, expecting '%s'", method),
				http.StatusMethodNotAllowed,
				w,
			)
			return
		}
		next(w, r)
	}
}
