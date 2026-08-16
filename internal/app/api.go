package app

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func NewApi() (huma.API, *http.ServeMux) {

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("My API", "1.0.0"))

	register(api, &deps{})

	return api, mux
}
