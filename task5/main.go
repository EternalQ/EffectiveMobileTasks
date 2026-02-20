package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

var ErrorDivisionByZero error = errors.New("Division by zero")

func main() {

}

func Division(a, b float32) (float32, error) {
	if b == 0 {
		return 0, ErrorDivisionByZero
	}

	return a / b, nil
}

func Sum(a ...int) int {
	if a == nil {
		return 0
	}

	res := 0
	for _, v := range a {
		res += v
	}
	return res
}

func SumHandler(w http.ResponseWriter, r *http.Request) {
	req := make([]int, 0)
	json.NewDecoder(r.Body).Decode(&req)

	res := Sum(req...)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}