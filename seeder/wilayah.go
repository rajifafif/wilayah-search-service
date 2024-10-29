package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"
	"github.com/typesense/typesense-go/v2/typesense/api/pointer"
)

type Document struct {
	CityID       string `json:"city_id"`
	CityName     string `json:"city_name"`
	CreatedAt    int64  `json:"created_at"`
	DeletedAt    *int64 `json:"deleted_at"`
	DistrictID   string `json:"district_id"`
	DistrictName string `json:"district_name"`
	FullName     string `json:"full_name"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Postal       string `json:"postal"`
	PostalID     string `json:"postal_id"`
	ProvinceID   string `json:"province_id"`
	ProvinceName string `json:"province_name"`
	UpdatedAt    int64  `json:"updated_at"`
	VillageID    string `json:"village_id"`
	VillageName  string `json:"village_name"`
}

const (
	batchSize      = 100 // Number of records to process in each batch
	collectionName = "villages"
	serverKey      = "http://localhost:8108"
	apiKey         = "WA5JCpQzlyAL3hdoWoYlxBQXLl8i0SgHNNoQLO7QnrHebSaL"
)

func main() {
	// Connect to SQLite database
	db, err := sql.Open("sqlite3", "wilayah.db")
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Typesense client setup
	client := typesense.NewClient(
		typesense.WithServer(serverKey),
		typesense.WithAPIKey(apiKey),
	)

	schema := &api.CollectionSchema{
		Name: collectionName,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "village_id", Type: "string"},
			{Name: "village_name", Type: "string"},
			{Name: "district_id", Type: "string"},
			{Name: "district_name", Type: "string"},
			{Name: "city_id", Type: "string"},
			{Name: "city_name", Type: "string"},
			{Name: "province_id", Type: "string"},
			{Name: "province_name", Type: "string"},
			{Name: "postal_id", Type: "string"},
			{Name: "postal", Type: "string"},
			{Name: "full_name", Type: "string"},
			{Name: "created_at", Type: "int64"},
		},
		DefaultSortingField: pointer.String("created_at"),
	}

	// Create the collection if it doesn't exist
	_, err = client.Collection(collectionName).Retrieve(context.Background())
	if err != nil {
		if err.Error() == "collection not found" {
			// Collection does not exist, create it
			_, err = client.Collections().Create(context.Background(), schema)
			if err != nil {
				log.Printf("failed to create collection: %v", err)
			} else {
				log.Printf("collection %s created.", collectionName)
			}
		} else {
			log.Fatalf("failed to check collection existence: %v", err)
		}
	} else {
		log.Printf("collection %s already exists, skipping creation.", collectionName)
	}

	// Pagination setup
	offset := 0
	for {
		// Query to retrieve villages with a limit and offset
		query := fmt.Sprintf(`
			SELECT 
				v.id AS village_id, 
				v.name AS village_name, 
				d.id AS district_id, 
				d.name AS district_name, 
				c.id AS city_id, 
				c.name AS city_name, 
				p.id AS province_id, 
				p.name AS province_name
			FROM villages v
			JOIN districts d ON v.district_id = d.id
			JOIN cities c ON d.city_id = c.id
			JOIN provinces p ON c.province_id = p.id
			LIMIT %d OFFSET %d`, batchSize, offset)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatalf("failed to query villages: %v", err)
		}

		// Check if there are no more rows
		if !rows.Next() {
			break
		}

		/// Prepare a slice to hold documents
		var documents []Document

		// Map data from joined tables to Document struct
		for rows.Next() {
			var doc Document
			if err := rows.Scan(&doc.VillageID, &doc.VillageName, &doc.DistrictID, &doc.DistrictName,
				&doc.CityID, &doc.CityName, &doc.ProvinceID, &doc.ProvinceName); err != nil {
				log.Fatalf("failed to scan row: %v", err)
			}
			doc.ID = doc.VillageID // Assuming ID for Document is the VillageID
			doc.FullName = fmt.Sprintf("%s, %s, %s, %s", doc.VillageName, doc.DistrictName, doc.CityName, doc.ProvinceName)

			doc.CreatedAt = time.Now().Unix()
			doc.UpdatedAt = time.Now().Unix()

			documents = append(documents, doc)

			// Upsert when the batch size is reached
			if len(documents) >= batchSize {
				// Upsert the batch of documents into Typesense
				for _, doc := range documents {
					_, err := client.Collection(collectionName).Documents().Upsert(context.Background(), doc)
					if err != nil {
						log.Printf("failed to upsert document %s: %v", doc.ID, err)
					} else {
						log.Printf("successfully upserted document %s", doc.ID)
					}
				}
				// Clear the documents slice for the next batch
				documents = documents[:0]
			}
		}

		// Upsert any remaining documents after the loop
		if len(documents) > 0 {
			for _, doc := range documents {
				_, err := client.Collection(collectionName).Documents().Upsert(context.Background(), doc)
				if err != nil {
					log.Printf("failed to upsert document %s: %v", doc.ID, err)
				} else {
					log.Printf("successfully upserted document %s", doc.ID)
				}
			}
		}

		offset += batchSize

	}

	fmt.Println("Seeding completed.")
}
