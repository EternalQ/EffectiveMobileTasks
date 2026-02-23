package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

var ErrDivisionByZero error = errors.New("division by zero")
var ErrWrongInput error = errors.New("wrong input")

func main() {

}

func Division(a, b float32) (float32, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}

	return a / b, nil
}

func Max(a []int) (int, error) {
	if len(a) == 0 {
		return 0, ErrWrongInput
	}

	res := a[0]
	for _, v := range a {
		if v > res {
			res = v
		}
	}
	return res, nil
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
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res := Sum(req...)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

type Clock interface {
	Now() time.Time
}

func GetSuffix(clock Clock) string {
	if clock.Now().Hour() > 12 {
		return "PM"
	}
	return "AM"
}
