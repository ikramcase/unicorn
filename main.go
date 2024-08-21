package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Capabilities = []string{""}

type uniCorn struct {
	ID           int
	Name         string
	Capabilities []string
}

var (
	unicornRequests = make(map[int]bool)
	mu              sync.Mutex
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func main() {

	//errorLOG.txt file is created to log all the errors
	file, err := os.OpenFile("errorLOG.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("please try later, errorLOG file unavailable")
		return
	}
	log.SetOutput(file)

	Capabilities = append(Capabilities, "super strong", "fullfill wishes", "fighting capabilities", "fly", "swim", "sing", "run", "cry", "change color", "talk", "dance", "code", "design", "drive", "walk", "talk chinese", "lazy")
	http.HandleFunc("/api/get-unicorn", GetUnicorn)
	http.HandleFunc("/api/get-store-data", GetStoreData)
	http.ListenAndServe(":8888", nil)
}

func GetUnicorn(w http.ResponseWriter, r *http.Request) {
	fmt.Println("processing new request..")

	//read in func
	names := readingFunc("petnames.txt")
	adj := readingFunc("adj.txt")

	//setting time duration
	sleep_time := time.Duration(rand.Intn(1000)) * time.Millisecond

	//getting amount value from url
	values := r.URL.Query()
	amount, _ := strconv.Atoi(values.Get("amount"))

	items := []uniCorn{}
	for j := 0; j < amount; j++ {
		//adding a request ID with the unicorn
		requestID := generateRequestID()
		mu.Lock()
		unicornRequests[requestID] = false
		mu.Unlock()
		fmt.Println("Unicorn requested. Request ID: ", requestID)
		//adding name
		name := adj[rand.Intn(1345)] + "-" + names[rand.Intn(5800)]
		cap1 := []string{}

		//adding capabilities
		uniqueCaps := make(map[string]bool)

		for i := 0; i < 3; {
			cap := Capabilities[rand.Intn(18)]

			// Check if the capability is unique
			if !uniqueCaps[cap] {
				cap1 = append(cap1, cap)
				uniqueCaps[cap] = true // Mark the capability as added
				i++                    // Increment the count only when a unique capability is added
			}
		}
		item := uniCorn{
			ID:           requestID,
			Name:         name,
			Capabilities: cap1,
		}
		items = append(items, item)
		time.Sleep(sleep_time)

		unicornData := fmt.Sprintf("ID: %d\nName: %s\nCapability: %s\n\n", item.ID, item.Name, item.Capabilities)
		err := writeFuncLIFO(unicornData)
		if err != nil {
			log.Println("data reading error from data/store.txt")
			errorResponse := ErrorResponse{
				Message: "An error occurred",
				Code:    http.StatusInternalServerError, // Set an appropriate HTTP status code
			}
			// Encode the error response as JSON
			responseJSON, err := json.Marshal(errorResponse)
			if err != nil {
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
				return
			}
			// Set the appropriate content type and HTTP status code
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(errorResponse.Code)

			// Write the error response to the response writer
			_, err = w.Write(responseJSON)
			if err != nil {
				log.Println("Failed to write error response:", err)
			}

			return
		}
	}

	//response output
	response, _ := json.Marshal(items)
	fmt.Println("Unicorn ready..")
	w.Header().Set("Content-Type", "application/json")

	w.Write(response)
}

func GetStoreData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Getting store data ..")

	//read in func
	storeData := readingFunc("store.txt")

	//response output
	response, _ := json.Marshal(storeData)
	fmt.Println("store data ready..")
	w.Header().Set("Content-Type", "application/json")

	w.Write(response)
}

func readingFunc(filename string) []string {
	fn, err := os.Open("data/" + filename)
	if err != nil {
		log.Println("reading func: please try later, unicorn factory unavailable")
		return nil
	}
	defer fn.Close()

	var names []string
	var scanner = bufio.NewScanner(fn)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line) // Trim leading and trailing spaces
		if trimmedLine != "" {
			names = append(names, trimmedLine)
		}
	}
	return names
}

func writeFuncLIFO(unicornData string) error {
	// Create or open the file for writing
	fileName := "store.txt"
	file, err := os.OpenFile("data/"+fileName, os.O_CREATE|os.O_WRONLY, 0644)
	fmt.Println("read file")
	if err != nil {
		log.Println("please try later, 1 unicorn store unavailable")
		return err
	}
	defer file.Close()

	// Move the file cursor to the beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Println("file cursor error !")
		return err
	}

	// Read the existing data
	existingData := readingFunc(fileName)
	if err != nil {
		log.Println("read all data error !")
		return err
	}

	// Combine the new data with the existing data
	dataToWrite := addStringAtStart(existingData, unicornData)
	fmt.Println("data to write: ", dataToWrite)

	// Truncate the file to the length of the combined data
	err = file.Truncate(int64(len(dataToWrite)))
	if err != nil {
		log.Println("please try later, truncate error")
		return err
	}

	// Write the combined data to the beginning of the file
	_, err = file.WriteAt([]byte(convertToBytes(dataToWrite)), 0)
	if err != nil {
		log.Println("please try later, write full data error ")
		return err
	}

	fmt.Printf("Unicorn data written to %s\n", fileName)
	return nil
}

func generateRequestID() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(10000)
}

func addStringAtStart(original []string, strToAdd string) []string {
	// Create a new slice with the string added at the start
	newSlice := append([]string{strToAdd}, original...)

	return newSlice
}

func convertToBytes(dataToWrite []string) []byte {
	// Join the strings into a single string, separated by a newline character
	joinedString := strings.Join(dataToWrite, "\n")

	// Convert the joined string into bytes
	bytesData := []byte(joinedString)

	return bytesData
}
