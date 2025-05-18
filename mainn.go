package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/golang-sql/civil"
	"github.com/joho/godotenv"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/trufnetwork/sdk-go/core/tnclient"
	"github.com/trufnetwork/sdk-go/core/types"
	"github.com/trufnetwork/sdk-go/core/util"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	privateKey := os.Getenv("PRIVATE_KEY")
	providerURL := os.Getenv("PROVIDER_URL")
	providerAddress := os.Getenv("PROVIDER_ADDRESS")

	if privateKey == "" || providerURL == "" || providerAddress == "" {
		log.Fatalf("Missing required environment variables.")
	}

	// Initialize TRUF Network Client
	ctx := context.Background()
	pk, err := crypto.Secp256k1PrivateKeyFromHex(privateKey)
	if err != nil {
		log.Fatalf("Failed to create private key: %v", err)
	}
	signer := &auth.EthPersonalSigner{Key: *pk}
	tnClient, err := tnclient.NewClient(ctx, providerURL, tnclient.WithSigner(signer))
	if err != nil {
		log.Fatalf("Failed to initialize TRUF network client: %v", err)
	}

	// Check node health
	//health, err := tnClient.Health(ctx)
	//if err != nil {
	//	log.Fatalf("Node health check failed: %v", err)
	//}
	//log.Printf("Node health: %+v", health)

	// Fetch and process inflation data
	inflationStreamID := util.GenerateStreamId("stf389ad7681059ca7750dda907735b2")
	log.Printf("Fetching inflation stream: %s with provider: %s", inflationStreamID, providerAddress)
	inflationData := fetchData(ctx, tnClient, inflationStreamID, providerAddress)
	fmt.Println("Inflation Data:")
	for _, record := range inflationData {
		fmt.Printf("Date (DateValue): %v, Value: %f\n", record["Date"], record["Value"])
	}

	// Fetch and process index data
	indexStreamID := util.GenerateStreamId("st15889445eac65d03159da7c882a895")
	log.Printf("Fetching index stream: %s with provider: %s", indexStreamID, providerAddress)
	indexData := fetchData(ctx, tnClient, indexStreamID, providerAddress)
	fmt.Println("Index Data:")
	for _, record := range indexData {
		fmt.Printf("Date (DateValue): %v, Index Value: %f\n", record["Date"], record["Value"])
	}

	// Calculate risk metrics and generate alerts for inflation data
	inflationRiskMetrics := calculateRiskMetrics(inflationData)
	fmt.Println("Inflation Risk Metrics:", inflationRiskMetrics)
	generateAlerts(inflationRiskMetrics)

	// Calculate risk metrics and generate alerts for index data
	indexRiskMetrics := calculateRiskMetrics(indexData)
	fmt.Println("Index Risk Metrics:", indexRiskMetrics)
	generateAlerts(indexRiskMetrics)
}

// Helper Function to Fetch and Process Data
func fetchData(ctx context.Context, tnClient *tnclient.Client, streamID util.StreamId, providerAddress string) []map[string]interface{} {
	dataProvider, err := util.NewEthereumAddressFromString(providerAddress)
	if err != nil {
		log.Fatalf("Invalid provider address: %v", err)
	}

	dateFrom := civil.Date{Year: 2023, Month: 1, Day: 1} // Adjust start date as needed
	dateTo := civil.Date{Year: 2023, Month: 1, Day: 31}  // Adjust end date as needed

	streamLocator := types.StreamLocator{
		StreamId:     streamID,
		DataProvider: dataProvider,
	}
	stream, err := tnClient.LoadPrimitiveStream(streamLocator)
	if err != nil {
		log.Fatalf("Failed to load stream %s: %v", streamID, err)
	}

	records, err := stream.GetRecord(ctx, types.GetRecordInput{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	})
	if err != nil {
		log.Fatalf("Failed to fetch records for stream %s: %v", streamID, err)
	}

	var processedRecords []map[string]interface{}
	for _, record := range records {
		value, err := record.Value.Float64()
		if err != nil {
			log.Printf("Non-finite value encountered. Skipping record: %+v", record)
			continue
		}
		scaledValue := value // Adjust scaling if needed
		log.Printf("Processed record: Date=%v, ScaledValue=%f", record.DateValue, scaledValue)
		processedRecords = append(processedRecords, map[string]interface{}{
			"Date":  record.DateValue,
			"Value": scaledValue,
		})
	}

	return processedRecords
}

// Helper Function to Calculate Risk Metrics
func calculateRiskMetrics(data []map[string]interface{}) map[string]float64 {
	riskMetrics := make(map[string]float64)

	var sum float64
	for _, record := range data {
		value := record["Value"].(float64)
		sum += value
	}
	if len(data) > 0 {
		riskMetrics["VaR"] = sum / float64(len(data)) // Simple average as VaR
	}

	return riskMetrics
}

// Helper Function to Generate Alerts
func generateAlerts(riskMetrics map[string]float64) {
	threshold := 1000.0 // Adjust threshold as needed
	if VaR, exists := riskMetrics["VaR"]; exists && VaR > threshold {
		fmt.Printf("ALERT: Portfolio Value-at-Risk exceeds threshold! VaR: %.2f\n", VaR)
	} else {
		log.Printf("No alert triggered. VaR: %.2f", riskMetrics["VaR"])
	}
}
