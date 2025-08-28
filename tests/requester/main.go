package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL = "http://localhost:9000/order/"
	fixedID = "M2fJlvVUpJY0ZvLY"
)

func main() {
	for {
		var wg sync.WaitGroup
		for range rand.Intn(10) {
			wg.Go(doRequest)
		}
		wg.Wait()
		time.Sleep(20 * time.Millisecond)
	}
}

func randomID(length int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	id := make([]rune, length)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return string(id)
}

func doRequest() {
	id := fixedID
	if rand.Intn(5) == 0 {
		id = randomID(12)
	}

	url := baseURL + id
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка запроса:", err)
	} else {
		fmt.Println("GET", url, "->", resp.Status)
		resp.Body.Close()
	}
}
