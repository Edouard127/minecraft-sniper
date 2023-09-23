package main

import (
	"context"
	"fmt"
	"github.com/Edouard127/go-mc/auth/data"
	"github.com/Edouard127/go-mc/auth/microsoft"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	go Ticker(context.Background(), time.Second*5, func() bool {
		InvokeDNSEntry(MojangRequest("/"))
		return false
	}) // Keep the DNS entry alive

	accounts, err := microsoft.ReadMinecraftAccounts()
	if err != nil {
		panic(err)
	}

	finder := NewThreeName(time.Minute * 5)

	for i := 0; i < len(accounts); i++ {
		wg.Add(1)
		go ClaimName(context.Background(), finder, func(index int, name string) bool { return index == i }, accounts[i], &wg)
	}

	wg.Wait()
	fmt.Println("Done")
}

func ClaimName(ctx context.Context, finder Finder, filter func(index int, name string) bool, account data.Auth, wg *sync.WaitGroup) {
	defer wg.Done()
	info := finder.GetByFilter(filter)
	if info == nil {
		return
	}

	if !info.Available {
		fmt.Println("Name is not available, waiting until it is")
		WaitUntil(ctx, info.First(), GetLatency())
		info = finder.GetByFilter(filter)
	}

	fmt.Println("Name is available")
	Ticker(ctx, time.Second, func() bool {
		available, err := account.NameAvailable(info.Username)
		if err != nil {
			panic(err)
		}

		if available {
			err = account.ChangeName(info.Username)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Account %s has claimed the name %s\n", account.Name, info.Username)
		}

		after := time.Now().After(info.Second())
		if after && !available {
			fmt.Printf("Name %s is no longer available\n", info.Username)
		}

		return available || after
	})
}

func Ticker(ctx context.Context, every time.Duration, f func() bool) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if f() {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
