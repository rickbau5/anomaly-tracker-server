package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	atp "github.com/rickbau5/anomaly-tracker-proto"
	"github.com/rickbau5/anomaly-tracker-server/cmd/internal/tracker"
)

type response struct {
	Error     string        `json:"error,omitempty"`
	Message   string        `json:"message,omitempty"`
	Anomaly   *atp.Anomaly  `json:"anomaly,omitempty"`
	Anomalies []atp.Anomaly `json:"anomalies,omitempty"`

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
		writeErrorResponse("Internal error", http.StatusInternalServerError, w)
		return
	}

	var protoAnomaly atp.Anomaly
	if r.Method != http.MethodGet {
		bs, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(bs))
		if err := jsonpb.Unmarshal(bytes.NewBuffer(bs), &protoAnomaly); err != nil {
			log.Println("Cannot unmarshal to Anomaly proto:", err)
		}
		bs, _ = json.MarshalIndent(protoAnomaly, "", "  ")
	}

	var resp response
	switch r.Method {
	case http.MethodPost:
		resp = addAnomaly(protoAnomaly, apiKey)
	case http.MethodDelete:
		resp = deleteAnomaly(protoAnomaly, apiKey)
	case http.MethodPatch:
		resp = updateAnomaly(protoAnomaly, apiKey)
	case http.MethodGet:
		resp = getAnomalies(apiKey)
	default:
		resp = response{
			Error:      "Unrecognized method: " + r.Method,
			StatusCode: http.StatusMethodNotAllowed,
		}
	}
	writeFullResponse(resp, w)
}

func addAnomaly(anomaly atp.Anomaly, apiKey tracker.APIKey) response {
	if err := tracker.AddAnomaly(anomaly, apiKey); err != nil {
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

func deleteAnomaly(anomaly atp.Anomaly, apiKey tracker.APIKey) response {
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

func updateAnomaly(anomaly atp.Anomaly, apiKey tracker.APIKey) response {
	updatedAnomaly, err := tracker.ModifyAnomaly(anomaly, apiKey)
	if err != nil {
		log.Println("Failed updating anomaly:", err.Error())
		return response{
			Error:      sanitizeError(err),
			StatusCode: http.StatusNotAcceptable,
		}
	}
	return response{
		StatusCode: http.StatusOK,
		Message:    "updated",
		Anomaly:    updatedAnomaly,
	}
}

func getAnomalies(apiKey tracker.APIKey) response {
	anomalies, err := tracker.GetAnomaliesByAPIKey(apiKey)
	if err != nil {
		log.Println("Failed getting anomalies:", err.Error())
		return response{
			Error:      sanitizeError(err),
			StatusCode: http.StatusNotAcceptable,
		}
	}
	log.Printf("Found %d anomalies for api key '%s'.\n", len(anomalies), apiKey.Key)
	return response{
		Anomalies:  anomalies,
		StatusCode: http.StatusOK,
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

func sanitizeResponse(res response) response {
	sanitized := res
	sanitizeAnomaly := func(anom atp.Anomaly) atp.Anomaly {
		return atp.Anomaly{
			Id:      anom.Id,
			System:  anom.System,
			Type:    anom.Type,
			Name:    anom.Name,
			Created: anom.Created,
		}
	}

	if sanitized.Anomaly != nil {
		cleanedAnomaly := sanitizeAnomaly(*res.Anomaly)
		sanitized.Anomaly = &cleanedAnomaly
	}
	if sanitized.Anomalies != nil {
		var anomalies []atp.Anomaly
		for _, anom := range sanitized.Anomalies {
			cleanedAnomaly := sanitizeAnomaly(anom)
			anomalies = append(anomalies, cleanedAnomaly)
		}
		sanitized.Anomalies = anomalies
	}
	return sanitized
}

func writeFullResponse(res response, w http.ResponseWriter) {
	cleanResponse := sanitizeResponse(res)
	bytes, err := json.MarshalIndent(cleanResponse, "", "    ")
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
