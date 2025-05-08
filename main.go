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

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatalf("PRIVATE_KEY not found in environment")
	}

	ctx := context.Background()

	pk, err := crypto.Secp256k1PrivateKeyFromHex(privateKey)
	if err != nil {
		log.Fatalf("Failed to create private key: %v", err)
	}
	signer := &auth.EthPersonalSigner{Key: *pk}

	tnClient, err := tnclient.NewClient(ctx, "https://staging.tsn.truflation.com", tnclient.WithSigner(signer))
	if err != nil {
		log.Fatalf("Failed to initialize TRUF network client: %v", err)
	}

	// Generate a Stream ID or use an existing one
	streamId := util.GenerateStreamId("stf37ad83c0b92c7419925b7633c0e62") // Replace with your stream name
	streamLocator := types.StreamLocator{
		StreamId: streamId,
		DataProvider: func() util.EthereumAddress {
			address, err := util.NewEthereumAddressFromString("0x4710a8d8f0d845da110086812a32de6d90d7ff5c")
			if err != nil {
				log.Fatalf("Failed to get address: %v", err)
			}
			return address
		}(),
	}

	// Load Primitive Actions
	primitiveActions, err := tnClient.LoadComposedStream(streamLocator)
	if err != nil {
		log.Fatalf("Failed to load primitive actions: %v", err)
	}

	// Fetch and display inflation data using timestamps
	dateFrom := civil.Date{Year: 2023, Month: 1, Day: 1}
	dateTo := civil.Date{Year: 2023, Month: 1, Day: 31}

	records, err := primitiveActions.GetRecord(ctx, types.GetRecordInput{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	})
	if err != nil {
		log.Fatalf("Failed to fetch inflation data: %v", err)
	}

	// Print the fetched data
	fmt.Println("Inflation Data:")
	for _, record := range records {
		fmt.Printf("Date (String): %d, Value: %s\n", record.DateValue, record.Value.String())
	}
}
