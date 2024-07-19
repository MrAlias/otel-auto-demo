package main

import (
	"context"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const scope = "github.com/MrAlias/otel-auto-demo/frontend"

func newServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()

	handle(mux, "/rolldice/", http.HandlerFunc(rolldice))
	handle(mux, "/rolldice/{player}", http.HandlerFunc(rolldice))

	return &http.Server{
		Addr:        addr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     otelhttp.NewHandler(mux, "/"),
	}
}

func handle(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, otelhttp.WithRouteTag(pattern, handler))
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	tracer := trace.SpanFromContext(r.Context()).TracerProvider().Tracer(scope)
	_, span := tracer.Start(r.Context(), "rolldice")
	defer span.End()

	roll := 1 + rand.Intn(6)

	if player := r.PathValue("player"); player != "" {
		span.SetAttributes(attribute.String("player", player))
	}
	span.SetAttributes(attribute.Int("roll.value", roll))

	resp := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		span.RecordError(err)
	}
}
