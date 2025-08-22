package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	humago "github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wiredmatt/go_short/internal/api/middleware"
	"github.com/wiredmatt/go_short/internal/shortener"
)

type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"ok"`
	}
	Status int `json:"status" example:"200"`
}

type ShortenInput struct {
	Body struct {
		UserID string `json:"userId"`
		URL    string `json:"url"`
	}
}
type ShortenOutput struct {
	Body struct {
		ShortURL string `json:"short_url"`
	}
	Status int `json:"status" example:"200"`
}

type ResolveInput struct {
	Code string `path:"code"`
}
type ResolveOutput struct {
	Location string `header:"Location"`
	Status   int    `json:"status" example:"302"`
}

func NewRouter(service shortener.Shortener) *http.ServeMux {
	apiMux := http.NewServeMux()

	// Initialize Huma on this mux
	humaAPI := humago.New(apiMux, huma.DefaultConfig("URL Shortener API", "1.0.0"))

	middleware.PrometheusInit()

	humaAPI.UseMiddleware(middleware.RequestID)
	humaAPI.UseMiddleware(middleware.RequestLogger)
	humaAPI.UseMiddleware(middleware.TrackMetrics)

	huma.Register(humaAPI, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/health",
		Summary: "Health check",
	}, func(ctx context.Context, _ *struct{}) (*HealthOutput, error) {
		res := &HealthOutput{}
		res.Body.Status = "ok"
		res.Status = http.StatusOK
		return res, nil
	})

	huma.Register(humaAPI, huma.Operation{
		Method:  http.MethodPost,
		Path:    "/shorten",
		Summary: "Create a shortened URL",
	}, func(ctx context.Context, in *ShortenInput) (*ShortenOutput, error) {
		code, err := service.Shorten(in.Body.UserID, in.Body.URL)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, err.Error())
		}
		var out ShortenOutput
		out.Body.ShortURL = service.GetBaseURL() + "/" + code
		out.Status = http.StatusOK
		return &out, nil
	})

	huma.Register(humaAPI, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/{code}",
		Summary: "Resolve a shortened URL",
	}, func(ctx context.Context, in *ResolveInput) (*ResolveOutput, error) {
		originalURL, err := service.Resolve(in.Code)
		if err != nil || originalURL == "" {
			return nil, huma.NewError(http.StatusNotFound, "not found")
		}
		return &ResolveOutput{
			Location: originalURL,
			Status:   http.StatusFound,
		}, nil
	})

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	root := http.NewServeMux()
	root.Handle("/metrics", promhttp.Handler())
	root.Handle("/", apiMux)

	return root
}
