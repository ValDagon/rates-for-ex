// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Struct for BTC to USD response from CoinDesk
type BTCToUSDResponse struct {
	Bpi struct {
		USD struct {
			Code        string  `json:"code"`
			Symbol      string  `json:"symbol"`
			Rate        string  `json:"rate"`
			Description string  `json:"description"`
			RateFloat   float64 `json:"rate_float"`
		} `json:"USD"`
	} `json:"bpi"`
}

// Struct to store data points
type DataPoint struct {
	Time     time.Time
	USDToEUR float64
	BTCToUSD float64
}

var (
	data   []DataPoint
	mu     sync.Mutex
	status string = "All is working" // Инициализируем статус
)

// Function to fetch BTC to USD from CoinDesk API
func fetchBTCToUSD() (float64, error) {
	url := "https://api.coindesk.com/v1/bpi/currentprice/USD.json"
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching BTC to USD: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("non-200 HTTP status for BTC to USD: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading BTC to USD response: %v", err)
	}

	var btcResp BTCToUSDResponse
	err = json.Unmarshal(body, &btcResp)
	if err != nil {
		return 0, fmt.Errorf("error parsing BTC to USD JSON: %v", err)
	}

	return btcResp.Bpi.USD.RateFloat, nil
}

// Function to fetch USD to EUR by scraping x-rates.com
func fetchUSDtoEUR() (float64, error) {
	url := "https://www.x-rates.com/calculator/?from=USD&to=EUR&amount=1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request for USD to EUR: %v", err)
	}
	// Используем реалистичный User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error fetching USD to EUR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("non-200 HTTP status for USD to EUR: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error parsing USD to EUR HTML: %v", err)
	}

	// Найдем <span class="ccOutputRslt">0.9425<span class="ccOutputTrail">86</span><span class="ccOutputCode"> EUR</span></span>
	rateStr := ""
	doc.Find("span.ccOutputRslt").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Основная часть курса
		mainRate := strings.TrimSpace(s.Contents().FilterFunction(func(i int, s *goquery.Selection) bool {
			return goquery.NodeName(s) == "#text"
		}).Text())

		// Дополнительная часть курса
		trailingRate := strings.TrimSpace(s.Find("span.ccOutputTrail").Text())

		if mainRate == "" && trailingRate == "" {
			return true // Продолжить поиск
		}

		if trailingRate == "" {
			rateStr = mainRate
		} else {
			rateStr = mainRate + trailingRate
		}
		return false // Прекратить поиск после первого совпадения
	})

	if rateStr == "" {
		return 0, fmt.Errorf("USD to EUR rate not found")
	}

	// Удаляем любые запятые и парсим число
	rateStr = strings.ReplaceAll(rateStr, ",", "")
	rateValue, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing USD to EUR rate: %v", err)
	}

	return rateValue, nil
}

// Function to fetch and store data
func fetchData() {
	currentTime := time.Now()

	usdToEur, err := fetchUSDtoEUR()
	if err != nil {
		log.Println("Error fetching USD to EUR:", err)
		mu.Lock()
		status = "Error fetching USD to EUR"
		mu.Unlock()
		return
	}

	btcToUsd, err := fetchBTCToUSD()
	if err != nil {
		log.Println("Error fetching BTC to USD:", err)
		mu.Lock()
		status = "Error fetching BTC to USD"
		mu.Unlock()
		return
	}

	mu.Lock()
	data = append(data, DataPoint{
		Time:     currentTime,
		USDToEUR: usdToEur,
		BTCToUSD: btcToUsd,
	})

	// Remove data older than 1 hour (изменено с 30 минут на 1 час)
	cutoff := currentTime.Add(-1 * time.Hour)
	i := 0
	for i < len(data) && data[i].Time.Before(cutoff) {
		i++
	}
	data = data[i:]
	status = "All is working"
	mu.Unlock()

	log.Printf("Fetched USD to EUR: %.6f at %s", usdToEur, currentTime.Format(time.RFC3339))
	log.Printf("Fetched BTC to USD: %.4f at %s", btcToUsd, currentTime.Format(time.RFC3339)) // 4 знака после запятой
}

// Goroutine to periodically fetch data every 10 seconds
func fetchExchangeRates() {
	fetchData()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		fetchData()
	}
}

// Handler for the main page
func handler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "index.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Println("Template parsing error:", err)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	times := []string{}
	usdToEur := []float64{}
	btcToUsd := []float64{}
	var currentUSDtoEUR, currentBTCtoUSD string

	for _, point := range data {
		times = append(times, point.Time.Format(time.RFC3339))
		usdToEur = append(usdToEur, point.USDToEUR)
		btcToUsd = append(btcToUsd, point.BTCToUSD)
	}

	if len(data) > 0 {
		last := data[len(data)-1]
		currentUSDtoEUR = fmt.Sprintf("%.6f", last.USDToEUR)
		currentBTCtoUSD = fmt.Sprintf("%.4f", last.BTCToUSD) // 4 знака после запятой
	}

	tmplData := map[string]interface{}{
		"Times":           template.JS(jsonMustMarshal(times)),
		"USDToEUR":        template.JS(jsonMustMarshal(usdToEur)),
		"BTCToUSD":        template.JS(jsonMustMarshal(btcToUsd)),
		"CurrentUSDtoEUR": currentUSDtoEUR,
		"CurrentBTCtoUSD": currentBTCtoUSD,
		"Status":          template.JS(jsonMustMarshal(status)),
	}

	err = t.Execute(w, tmplData)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
		return
	}
}

// Handler for the /data endpoint
func dataHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	type DataResponse struct {
		Times           []string  `json:"times"`
		USDToEUR        []float64 `json:"usdToEur"`
		BTCToUSD        []float64 `json:"btcToUsd"`
		CurrentUSDtoEUR string    `json:"currentUSDtoEUR"`
		CurrentBTCtoUSD string    `json:"currentBTCtoUSD"`
		Status          string    `json:"status"`
	}

	times := []string{}
	usdToEur := []float64{}
	btcToUsd := []float64{}
	var currentUSDtoEUR, currentBTCtoUSD, currentStatus string

	for _, point := range data {
		times = append(times, point.Time.Format(time.RFC3339))
		usdToEur = append(usdToEur, point.USDToEUR)
		btcToUsd = append(btcToUsd, point.BTCToUSD)
	}

	if len(data) > 0 {
		last := data[len(data)-1]
		currentUSDtoEUR = fmt.Sprintf("%.6f", last.USDToEUR)
		currentBTCtoUSD = fmt.Sprintf("%.4f", last.BTCToUSD) // 4 знака после запятой
	}

	currentStatus = status

	response := DataResponse{
		Times:           times,
		USDToEUR:        usdToEur,
		BTCToUSD:        btcToUsd,
		CurrentUSDtoEUR: currentUSDtoEUR,
		CurrentBTCtoUSD: currentBTCtoUSD,
		Status:          currentStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Println("JSON encoding error:", err)
		return
	}
}

// Helper function to marshal JSON and handle errors
func jsonMustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return "[]"
	}
	return string(b)
}

func main() {
	// Start fetching exchange rates
	go fetchExchangeRates()

	// Handle graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Set up HTTP handlers
	http.HandleFunc("/", handler)
	http.HandleFunc("/data", dataHandler)

	srv := &http.Server{
		Addr: ":8080",
	}

	go func() {
		log.Println("Server started on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %s", err)
		}
	}()

	// Wait for termination signal
	<-sigs
	log.Println("Shutting down server...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server gracefully stopped")
}
