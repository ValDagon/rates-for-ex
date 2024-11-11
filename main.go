package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// API response structures
type USDToEURResponse struct {
	Rates struct {
		EUR float64 `json:"EUR"`
	} `json:"rates"`
	Result string `json:"result"`
}

type BTCToUSDResponse struct {
	Bpi struct {
		USD struct {
			RateFloat float64 `json:"rate_float"`
		} `json:"USD"`
	} `json:"bpi"`
}

type DataPoint struct {
	Time     string
	USDToEUR float64
	BTCToUSD float64
}

var (
	data []DataPoint
	mu   sync.Mutex
)

// Function to fetch data from API
func fetchAPI(ctx context.Context, url string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

// Function to periodically fetch exchange rates
func fetchExchangeRates() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		currentTime := time.Now().Format("15:04:05")

		// Fetch USD to EUR rate
		var usdToEurResp USDToEURResponse
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := fetchAPI(ctx, "https://open.er-api.com/v6/latest/USD", &usdToEurResp)
		cancel()
		if err != nil {
			log.Println("Error fetching USD to EUR:", err)
			continue
		}

		// Fetch BTC to USD rate
		var btcToUsdResp BTCToUSDResponse
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		err = fetchAPI(ctx, "https://api.coindesk.com/v1/bpi/currentprice/USD.json", &btcToUsdResp)
		cancel()
		if err != nil {
			log.Println("Error fetching BTC to USD:", err)
			continue
		}

		// Save data point
		mu.Lock()
		data = append(data, DataPoint{
			Time:     currentTime,
			USDToEUR: usdToEurResp.Rates.EUR,
			BTCToUSD: btcToUsdResp.Bpi.USD.RateFloat,
		})

		// Keep only the last 20 data points
		if len(data) > 20 {
			data = data[len(data)-20:]
		}
		mu.Unlock()
	}
}

// HTTP handler function
func handler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Prepare data for the template
	times := []string{}
	usdToEur := []float64{}
	btcToUsd := []float64{}
	for _, point := range data {
		times = append(times, point.Time)
		usdToEur = append(usdToEur, point.USDToEUR)
		btcToUsd = append(btcToUsd, point.BTCToUSD)
	}

	// Serialize data to JSON
	timesJSON, err := json.Marshal(times)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	usdToEurJSON, err := json.Marshal(usdToEur)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	btcToUsdJSON, err := json.Marshal(btcToUsd)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	tmplData := map[string]interface{}{
		"Times":    template.JS(timesJSON),
		"USDToEUR": template.JS(usdToEurJSON),
		"BTCToUSD": template.JS(btcToUsdJSON),
	}

	// Load template from file
	tmplPath := filepath.Join("templates", "index.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Println("Error loading template:", err)
		return
	}

	err = t.Execute(w, tmplData)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
		return
	}
}

func main() {
	// Start fetching exchange rates
	go fetchExchangeRates()

	// Handle system signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Setup and start HTTP server
	http.HandleFunc("/", handler)
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

	// Create context with timeout for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server gracefully stopped")
}
