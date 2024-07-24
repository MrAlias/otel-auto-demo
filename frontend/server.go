package main

import (
	"context"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/MrAlias/otel-auto-demo/quota"
	"github.com/MrAlias/otel-auto-demo/user"
)

const scope = "github.com/MrAlias/otel-auto-demo/frontend"

type serviceKeyType int

const (
	userKey serviceKeyType = iota
	quotaKey
)

func newServer(ctx context.Context, listenAddr, userAddr, quotaAddr string) *http.Server {
	uClient := user.NewClient(otelhttp.DefaultClient, userAddr)
	if err := uClient.HealthCheck(ctx); err != nil {
		log.Print("Cannot reach User service: ", err)
	} else {
		log.Print("Connected to User service at ", userAddr)
	}
	ctx = context.WithValue(ctx, userKey, uClient)

	qClient := quota.NewClient(otelhttp.DefaultClient, quotaAddr)
	if err := qClient.HealthCheck(ctx); err != nil {
		log.Print("Cannot reach Quota service: ", err)
	} else {
		log.Print("Connected to Quota service at ", quotaAddr)
	}
	ctx = context.WithValue(ctx, quotaKey, qClient)

	mux := http.NewServeMux()

	handle(mux, "/rolldice/", http.HandlerFunc(rolldice))
	handle(mux, "/rolldice/{player}", http.HandlerFunc(rolldice))

	return &http.Server{
		Addr:        listenAddr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     otelhttp.NewHandler(mux, "/"),
	}
}

func handle(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, otelhttp.WithRouteTag(pattern, handler))
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer(scope)
	_, span := tracer.Start(r.Context(), "rolldice")
	defer span.End()

	player := r.PathValue("player")

	uClient, ok := ctx.Value(userKey).(*user.Client)
	if !ok {
		http.Error(w, "Internal Error", http.StatusInternalServerError)

		err := errors.New("no User client")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	qClient, ok := ctx.Value(quotaKey).(*quota.Client)
	if !ok {
		http.Error(w, "Internal Error", http.StatusInternalServerError)

		err := errors.New("no Quota client")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	id, err := uClient.UserID(ctx, player)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	if err := qClient.UseQuota(ctx, id); err != nil {
		if errors.Is(err, quota.ErrInsufficient) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		} else {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	roll := 1 + rand.Intn(6)

	if player != "" {
		span.SetAttributes(attribute.String("player", player))
	}
	span.SetAttributes(attribute.Int("roll.value", roll))

	resp := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
