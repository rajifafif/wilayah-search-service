package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"
)

// Pointer utilities
func String(s string) *string {
	return &s
}

func Int64(i int64) *int64 {
	return &i
}

func Int(i int) *int {
	return &i
}

const (
	collectionName = "villages"
	serverKey      = "http://localhost:8108"
	apiKey         = "WA5JCpQzlyAL3hdoWoYlxBQXLl8i0SgHNNoQLO7QnrHebSaL"
	servicePort    = ":8109"
)

func getAllWilayahs(c echo.Context) error {
	client := typesense.NewClient(
		// typesense.WithServer("http://localhost:8108"),
		typesense.WithServer(serverKey),
		typesense.WithAPIKey(apiKey),
	)

	// Get search query from the request parameters
	query := c.QueryParam("q")
	queryBy := c.QueryParam("query_by")
	perPageStr := c.QueryParam("per_page")

	// Set default value for queryBy if it's not provided
	if queryBy == "" {
		queryBy = "full_name"
	}

	PerPage := 10 // default value
	if perPageStr != "" {
		if parsedPerPage, err := strconv.Atoi(perPageStr); err == nil {
			PerPage = parsedPerPage
		}
	}

	// Prepare search parameters
	searchParameters := &api.SearchCollectionParams{
		Q:       String(query),
		QueryBy: String(queryBy),
		PerPage: Int(PerPage), // Set the number of results per page
	}

	// Execute the search
	searchResults, err := client.Collection(collectionName).Documents().Search(context.Background(), searchParameters)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to search"})
	}

	return c.JSON(http.StatusOK, searchResults)
}

func main() {
	e := echo.New()

	// Middleware
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/wilayah", getAllWilayahs)

	// Start server
	e.Logger.Fatal(e.Start(servicePort))
}
