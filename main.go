package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type RequestBody struct {
	Sum              string  `json:"sum"`
	CurrencyCode     string  `json:"currency_code"`
	PaymentSystemKey string  `json:"payment_system_key"`
	Details          Details `json:"details"`
}

type Details struct {
	MerchantID string `json:"merchant_id"`
	OrderID    string `json:"order_id"`
}

type ResponseBody struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	ID                 string         `json:"id"`
	Sum                string         `json:"sum"`
	CurrencyCode       string         `json:"currency_code"`
	PaymentSystemKey   string         `json:"payment_system_key"`
	PaymentDetails     PaymentDetails `json:"payment_details"`
	PaymentDetailsLink string         `json:"payment_details_link"`
	PaymentDetailsQR   string         `json:"payment_details_qr"`
	RedirectURL        string         `json:"redirect_url"`
}

type PaymentDetails struct {
	Phone string `json:"phone"`
	UPIID string `json:"upi_id"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/api/payment", handlePaymentRequest)
	http.ListenAndServe(":8080", nil)
}

func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody RequestBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	client := &http.Client{}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Failed to convert request body", http.StatusInternalServerError)
		return
	}

	requestBodyReader := strings.NewReader(string(requestBodyBytes))

	apiBaseUrl := os.Getenv("API_BASE_URL")
	apiKey := os.Getenv("API_KEY")

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/merchant/api/order/in/init?key=%s", apiBaseUrl, apiKey),
		requestBodyReader,
	)

	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	var responseBody ResponseBody
	err = json.Unmarshal(respBody, &responseBody)
	if err != nil {
		http.Error(w, "Failed to parse response body", http.StatusInternalServerError)
		return
	}

	response := struct {
		RedirectURL string `json:"redirect_url"`
	}{
		RedirectURL: responseBody.Data.RedirectURL,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
