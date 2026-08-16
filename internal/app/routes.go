package app

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type Out struct {
	Body struct {
		Value string `json:"value"`
	}
}

func register(api huma.API, deps *deps) {
	huma.Get(api, "/greetings/{name}", func(ctx context.Context, i *struct {
		Name string `path:"name"`
	}) (*Out, error) {
		resp := &Out{}
		resp.Body.Value = "siemanko, " + i.Name
		return resp, nil
	})
}
