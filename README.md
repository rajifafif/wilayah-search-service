
# Wilayah Text Search

API Service for indonesian wilayah, using Typesense to search.

Using data from https://github.com/cahyadsn/wilayah 
`db/wilayah.sql (data kode wilayah terbaru sesuai dengan Kepmendagri No 100.1.1-6117 Tahun 2022)`

## To Deploy
### 1. Install [Typesense](https://typesense.org/) and run the service. 

Copy the `api-key` from `/etc/typesense/typesense-server.ini`

### 2. Seeding data
- Change API Key and URL in `seeder/wilayah.go`
```
const (
	batchSize      = 100 // Number of records to process in each batch
	collectionName = "villages"
	serverKey      = "http://localhost:8108"
	apiKey         = "YOURTYPESENSEAPIKEY"
)
```
- Process to upsert the wilayah data
```go
go run seeder/wilayah.go
```
This will read from `wilayah.db` which are sqlite data from `wilayah.sql`

### 3. Run the Service
```
go run server.go
```
This command will run the server based on the const configuration
```
const (
	collectionName = "villages"
	serverKey      = "http://localhost:8108"
	apiKey         = "YOURTYPESENSEAPIKEY"
	servicePort    = ":8109"
)
```
go make some coffee because this current upsert methods takes too long to finish

## Usage/Examples


`http://localhost:8109/wilayah?q=bergas jawa tengah=query&query_by=full_name&per_page=10`

option on query_by 
- full_name
- city_name
- province_name
- district_name
- village_name

but the response are villages resource, not province/city/district

