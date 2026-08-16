package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gbujak/reason-and-dignity/m/v2/internal/app"
)

func main() {
	api, _ := app.NewApi()
	spec := api.OpenAPI()

	bytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("openapi.json", bytes, 0666); err != nil {
		log.Fatal(err)
	}
}
