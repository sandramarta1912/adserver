package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type targetingInfo struct {
	IP      string
	Countries []string
}

type partner struct {
	Id      string        `db:"id"`
	IsSsp   bool          `db:"is_ssp"`
	IsDsp   bool          `db:"is_dsp"`
	Name    string        `db:"name"`
	Timeout time.Duration `db:"timeout"`
	URL     string        `db:"url"`
	Method  string        `db:"method"`
}

type bid struct {
	Id        string
	URL       string
	Value     float64
	PartnerId string
}

func AdServerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		//_ = r.URL.Query()["ip"][0]
		//_ = r.URL.Query()["country"][0]
	}
	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var targetingInfo targetingInfo
		err = json.Unmarshal(body, &targetingInfo)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Print(targetingInfo)
	}

	p := partner{"p1", true, false, "partner1", 200, "", ""}

	partners := []partner{{"dp1", false, true, "dpartner1", 0, "http://localhost:3002", "GET"},
		{"dp2", false, true, "dpartner2", 0, "http://localhost:3002", "GET"},
		{"dp3", false, true, "dpartner3", 0, "http://localhost:3002", "GET"},
		{"dp4", false, true, "dpartner4", 0, "http://localhost:3002", "GET"}}

	ctx, cancel := context.WithTimeout(context.Background(), p.Timeout*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(partners) + 1)

	bids := make([]bid, 0)
	receivedBids := make(chan bid, len(partners))

	go func() {
		defer wg.Done()
		for range partners {
			select {
			case a := <-receivedBids:
				bids = append(bids, a)
			case <-ctx.Done():
				fmt.Println(ctx.Err())
				return
			}
		}
	}()

	for _, p := range partners {
		go func() {
			defer wg.Done()
			select {
			case a := <-MakeRequest(p.URL, p.Method, p.Timeout):
				receivedBids <- a
			case <-ctx.Done():
				fmt.Println(ctx.Err())
			}
		}()
	}
	wg.Wait()
	bestBid := Max(bids)

	bidJson, err := json.Marshal(bestBid)
	if err != nil {
		panic(err)
	}
	w.Write(bidJson)
}

func Max(bids []bid) bid {
	max := 0.0
	for _, v := range bids {
		if v.Value > max {
			max = v.Value
		}
	}
	var bestBid bid
	for _, v := range bids {
		if v.Value == max {
			bestBid = v
		}
	}
	return bestBid
}

func MakeRequest(urlStr, method string, timeout time.Duration) chan bid {
	r, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{Timeout: timeout * time.Millisecond}
	resp, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	var receivedBid bid
	responseBody, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(responseBody, &receivedBid)
	if err != nil {
		fmt.Print(err)
	}
	c := make(chan bid, 1)
	c <- receivedBid

	return c
}
