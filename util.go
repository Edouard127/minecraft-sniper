package main

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var MojangRequest = func(path string) string {
	return "https://api.minecraftservices.com" + path
}

func nopanic(a ...any) {
	for i := range a {
		if err, ok := a[i].(error); ok {
			if err != nil {
				panic(err)
			}
		}
	}
}

func InvokeDNSEntry(host string) {
	nopanic(net.LookupHost(strings.TrimSuffix(strings.SplitAfterN(host, "/", 4)[2], "/")))
}

func CreateRequest(method, url string) (*http.Request, error) {
	return http.NewRequest(method, url, nil)
}

func Get(url string) (*http.Response, error) {
	req, err := CreateRequest("GET", url)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func GetDocumentFromURL(url string) (*goquery.Document, error) {
	resp, err := Get(url)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func ParseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}

	t, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.UnixMilli(t)
}

func GetLatency() time.Duration {
	req, err := CreateRequest("GET", MojangRequest("/"))
	if err != nil {
		return 0
	}

	var lowest time.Duration
	var result httpstat.Result
	ctx := httpstat.WithHTTPStat(req.Context(), &result)
	req = req.WithContext(ctx)

	for i := 0; i < 20; i++ {
		http.DefaultClient.Do(req)
		t := result.DNSLookup + result.TCPConnection + result.TLSHandshake + result.ServerProcessing
		if t < lowest {
			lowest = t
		}
	}
	return lowest
}

func WaitUntil(ctx context.Context, at time.Time, latency time.Duration) {
	timer := time.NewTimer(time.Until(at.Add(-latency)))
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		return
	}
}
