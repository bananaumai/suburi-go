package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
)

func main() {
	var prefix string
	flag.StringVar(&prefix, "p", "/", "prefix")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{"localhost:2379"},
		DialTimeout:          5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer func() { _ = cli.Close() }()

	readyToScan := make(chan struct{})
	errgrp, ctx := errgroup.WithContext(ctx)

	errgrp.Go(func() error {
		s := bufio.NewScanner(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("context is canceled; stop writing etcd\n")
				return nil
			case <-readyToScan:
			}
			var key, value string
			fmt.Printf("key:")
			if s.Scan() {
				key = s.Text()
			}

			fmt.Printf("value:")
			if s.Scan() {
				value = s.Text()
			}

			fmt.Printf("put %s: %s\n", key, value)
			if _, err := cli.Put(ctx, prefix+key, value); err != nil {
				return err
			}
			fmt.Printf("successfully put\n")
		}
	})

	errgrp.Go(func() error {
		watchCh := cli.Watch(ctx, prefix, clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("context is canceled; stop reading etcd\n")
				return nil
			default:
			}

			ctx, cancel := context.WithTimeout(ctx, time.Second)
			res, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
			if err != nil {
				cancel()
				return err
			}

			for _, kv := range res.Kvs {
				fmt.Printf("%s: %s\n", string(kv.Key), string(kv.Value))
			}
			cancel()

			readyToScan <- struct{}{}

			<-watchCh
			fmt.Printf("new kv detected\n")
		}
	})

	errs := make(chan error)
	go func() {
		if err := errgrp.Wait(); err != nil {
			errs <- err
		}
		errs <- nil
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
	case err := <-errs:
			fmt.Printf("error: %s", err)
	}
	cancel()
}
