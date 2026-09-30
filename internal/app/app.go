package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func Run() {
	demo := env("DEMO_MODE", "false") == "true"
	dir := env("DATA_DIR", "data/live")
	if demo {
		dir = env("DATA_DIR", "data/demo")
	}
	a, e := newApp(dir, demo)
	if e != nil {
		log.Fatal(e)
	}
	if !demo && a.token == "" {
		log.Fatal("Укажите MAX_BOT_TOKEN или включите DEMO_MODE=true")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go a.deliverNotifications(ctx)
	if env("BOT_MODE", "off") == "polling" {
		go a.poll(ctx)
	}
	srv := &http.Server{Addr: env("LISTEN_ADDR", "127.0.0.1:8080"), Handler: a.routes(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 90 * time.Second}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	fmt.Println("Docspace: http://" + srv.Addr)
	if e = srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		log.Fatal(e)
	}
}
