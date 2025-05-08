package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
	streamId := util.GenerateStreamId("TRUUK") // Replace with your stream name
	streamLocator := tnClient.OwnStreamLocator(streamId)

	// Load Primitive Actions
	primitiveActions, err := tnClient.LoadPrimitiveActions()
	if err != nil {
		log.Fatalf("Failed to load primitive actions: %v", err)
	}

	// Fetch and display inflation data using timestamps
	dateFrom := int(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Unix()) // Convert date to UNIX timestamp
	dateTo := int(time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC).Unix())

	records, err := primitiveActions.GetRecord(ctx, types.GetRecordInput{
		DataProvider: streamLocator.DataProvider.Address(), 
		StreamId:     streamLocator.StreamId.String(),      
		From:         &dateFrom,                            
		To:           &dateTo,                            
	})
	if err != nil {
		log.Fatalf("Failed to fetch inflation data: %v", err)
	}

	// Print the fetched data
	fmt.Println("Inflation Data:")
	for _, record := range records {
		fmt.Printf("Date (Timestamp): %d, Value: %s\n", record.EventTime, record.Value.String())
	}
}
