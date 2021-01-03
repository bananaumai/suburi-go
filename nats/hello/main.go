package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

const (
	publishInterval = 2 * time.Second
)

var (
	targets = []string{"banana", "apple", "orange"}
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	wg := sync.WaitGroup{}
	wg.Add(1)
	eg.Go(func() error {
		defer func() { wg.Done() }()
		if err := publishLoop(ctx, nc); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	wg.Add(1)
	eg.Go(func() error {
		defer func() { wg.Done() }()
		subs := make([]*nats.Subscription, len(targets))
		for i, t := range targets {
			subject := fmt.Sprintf("/hello/%s", t)
			sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
				log.Printf("received msg at %s, %s", msg.Subject, string(msg.Data))
			})
			if err != nil {
				return fmt.Errorf("subscribing error: %w", err)
			}
			subs[i] = sub
			log.Printf("subscribed %s", sub.Subject)
		}

		select {
		case <-ctx.Done():
		}

		for _, sub := range subs {
			if err := sub.Unsubscribe(); err != nil {
				log.Printf("failed to unsubscribe: %s", err)
			} else {
				log.Printf("unsubscribed %s", sub.Subject)
			}
		}

		return nil
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	errs := make(chan error, 1)
	go func() {
		err := eg.Wait()
		if err != nil {
			errs <- err
		}
	}()

	select {
	case <-sigs:
		cancel()
	case err := <-errs:
		panic(err)
	}

	wg.Wait()
	log.Printf("done")
}

func publishLoop(ctx context.Context, nc *nats.Conn) error {
	ticker := time.NewTicker(publishInterval)
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("publish loop is canceled: %w", ctx.Err())
		case <-ticker.C:
		}

		idx := rnd.Int() % len(targets)
		subject := fmt.Sprintf("/hello/%s", targets[idx])
		msg := []byte("hello")
		if err := nc.Publish(subject, msg); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
		log.Printf("published to %s", subject)
	}
}
