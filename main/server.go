package main

import "fmt"
import "net/http"
import "log"
import "encoding/json"
import "io/ioutil"
import "os"
import "github.com/joho/godotenv"

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func main() {
    http.HandleFunc("/api/devices", enableCORS(deviceHandler))
    http.HandleFunc("/api/preferences", enableCORS(preferenceHandler))

    fmt.Println("Server listening on port 3000")
    log.Panic(
        http.ListenAndServe(":3000", nil),
    )
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        next(w, r)
    }
}

func checkError(err error) {
    if err != nil {
        log.Panic(err)
    }
}

func deviceHandler(w http.ResponseWriter, r *http.Request) {

    req, err := http.NewRequest("GET", "https://track.onestepgps.com/v3/api/public/device", nil)

    q := req.URL.Query()
    apiKey := os.Getenv("ONESTEPGPS_API_KEY")
    q.Add("api-key", apiKey)
    q.Add("latest_point", "true")
    req.URL.RawQuery = q.Encode()

    client := &http.Client{}
    response, err := client.Do(req)
    checkError(err)

    defer response.Body.Close();

    type devicePoint struct {
        Lat float64 `json:"lat"`
        Lng float64 `json:"lng"`
    }

    type device struct {
        DeviceID string `json:"device_id"`
        LatestDevicePoint devicePoint `json:"latest_device_point"`
        DisplayName string `json:"display_name"`
        ActiveState string `json:"active_state"`
    }

    type deviceAPIResponse struct {
        ResultList []device `json:"result_list"`
    }

    var apiResponse deviceAPIResponse
    err = json.NewDecoder(response.Body).Decode(&apiResponse)
    checkError(err)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(apiResponse)
}

type userPreferences struct {
    Sort           string `json:"sort"`
    Highlight      []string `json:"highlight"`
}

func preferenceHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
        case http.MethodPost:

            var newSettings userPreferences
            err := json.NewDecoder(r.Body).Decode(&newSettings)
            if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusCreated)

            jsonString, _ := json.Marshal(newSettings)
            ioutil.WriteFile("settings.json", jsonString, os.ModePerm)

        case http.MethodGet:
            var preferences userPreferences

            jsonFile, err := os.Open("settings.json")
            if err != nil {
                fmt.Println(err)
            }

            if jsonFile != nil {
                fmt.Println("Successfully Opened settings.json")
                defer jsonFile.Close()

                byteValue, _ := ioutil.ReadAll(jsonFile)
                err = json.Unmarshal(byteValue, &preferences)
                if err != nil {
                    log.Fatalf("Unable to marshal JSON due to %s", err)
                }
            } else {
                preferences.Sort = "a-z"
                preferences.Highlight = nil
            }            

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(preferences)
    }
}
