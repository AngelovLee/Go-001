package main

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	server http.Server
}

func (s *Server) run() error {
	return s.server.ListenAndServe()
}

func (s *Server) shutDown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewServer(addr string) *Server {
	return &Server{
		server: http.Server{
			Addr: addr,
		},
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, _ := errgroup.WithContext(ctx)

	// run s1
	s1 := NewServer(":8000")
	g.Go(func() error {
		return s1.run()
	})

	// run s2
	s2 := NewServer(":8001")
	g.Go(func() error {
		return s2.run()
	})

	// watcher
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGPIPE)

	g.Go(func() error {
		for {
			select {
			case sig := <-quit:
				cancel()
				return errors.New("exit signal " + sig.String())
			case <-ctx.Done():
				signal.Stop(quit)
				s1.server.Shutdown(context.Background())
				s2.server.Shutdown(context.Background())
				log.Printf("shutDown signal")
				return nil
			}
		}
	})
	if err := g.Wait(); err != nil {
		cancel()
		log.Printf("Error:%v\n", err)
	}
}
