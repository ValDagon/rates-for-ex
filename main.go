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

// Структуры для данных API
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

// Функция для получения данных из API
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

// Функция для периодического получения курсов валют
func fetchExchangeRates() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		currentTime := time.Now().Format("15:04:05")

		// Получаем курс USD к EUR
		var usdToEurResp USDToEURResponse
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := fetchAPI(ctx, "https://open.er-api.com/v6/latest/USD", &usdToEurResp)
		cancel()
		if err != nil {
			log.Println("Ошибка получения USD к EUR:", err)
			continue
		}

		// Получаем курс BTC к USD
		var btcToUsdResp BTCToUSDResponse
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		err = fetchAPI(ctx, "https://api.coindesk.com/v1/bpi/currentprice/USD.json", &btcToUsdResp)
		cancel()
		if err != nil {
			log.Println("Ошибка получения BTC к USD:", err)
			continue
		}

		// Сохраняем точку данных
		mu.Lock()
		data = append(data, DataPoint{
			Time:     currentTime,
			USDToEUR: usdToEurResp.Rates.EUR,
			BTCToUSD: btcToUsdResp.Bpi.USD.RateFloat,
		})

		// Оставляем последние 20 точек данных
		if len(data) > 20 {
			data = data[len(data)-20:]
		}
		mu.Unlock()
	}
}

// Обработчик HTTP-запросов
func handler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Преобразуем данные для шаблона
	times := []string{}
	usdToEur := []float64{}
	btcToUsd := []float64{}
	for _, point := range data {
		times = append(times, point.Time)
		usdToEur = append(usdToEur, point.USDToEUR)
		btcToUsd = append(btcToUsd, point.BTCToUSD)
	}

	// Сериализуем данные в JSON
	timesJSON, err := json.Marshal(times)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	usdToEurJSON, err := json.Marshal(usdToEur)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	btcToUsdJSON, err := json.Marshal(btcToUsd)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	tmplData := map[string]interface{}{
		"Times":    template.JS(timesJSON),
		"USDToEUR": template.JS(usdToEurJSON),
		"BTCToUSD": template.JS(btcToUsdJSON),
	}

	// Загружаем шаблон из файла
	tmplPath := filepath.Join("templates", "index.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		log.Println("Ошибка загрузки шаблона:", err)
		return
	}

	err = t.Execute(w, tmplData)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		log.Println("Ошибка выполнения шаблона:", err)
		return
	}
}

func main() {
	// Запускаем сбор курсов валют
	go fetchExchangeRates()

	// Обрабатываем системные сигналы для корректного завершения работы
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Настраиваем и запускаем HTTP-сервер
	http.HandleFunc("/", handler)
	srv := &http.Server{
		Addr: ":8080",
	}

	go func() {
		log.Println("Сервер запущен на порту 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %s", err)
		}
	}()

	// Ожидаем сигнала завершения
	<-sigs
	log.Println("Завершение работы сервера...")

	// Создаем контекст с тайм-аутом для завершения сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении сервера: %v", err)
	}
	log.Println("Сервер успешно завершил работу")
}
