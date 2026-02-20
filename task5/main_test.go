package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestDivision(t *testing.T) {
	tests := []struct {
		name string

		a       float32
		b       float32
		want    float32
		wantErr error
	}{
		{
			name:    "Simple division",
			a:       10.0,
			b:       2.0,
			want:    5.0,
			wantErr: nil,
		},
		{
			name:    "Division with error",
			a:       10.0,
			b:       0.0,
			want:    0.0,
			wantErr: ErrorDivisionByZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := Division(tt.a, tt.b)
			if tt.wantErr != nil {
				assert.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				assert.Nil(t, gotErr)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name string

		a    []int
		want int
	}{
		{
			name: "Zero sum",
			a:    make([]int, 0),
			want: 0,
		},
		{
			name: "Nil input",
			a:    nil,
			want: 0,
		},
		{
			name: "Normal sum",
			a:    []int{1, 2, 3, 4, 5},
			want: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sum(tt.a...)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDivisionProperty(t *testing.T) {
	prop := func(a, b float32) bool {
		if b == 0 {
			return true
		}

		res, _ := Division(a, b)
		if (a >= 0 && b > 0) || (a <= 0 && b < 0) {
			return res >= 0
		} else {
			return res <= 0
		}
	}

	if err := quick.Check(prop, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Sign property error: %v", err)
	}
}
func TestSumHandler(t *testing.T) {
	payload := []int{1, 2, 3, 4, 5}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	SumHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var got int
	json.NewDecoder(rr.Body).Decode(&got)
	want := 15

	if got != want {
		t.Errorf("unexpected body: got %v want %v", got, want)
	}
}

func TestCounter_Increment(t *testing.T) {
	counter := &Counter{}
	var wg sync.WaitGroup
	numGoroutines := 100
	increments := 100

	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range increments {
				counter.Increment()
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, numGoroutines*increments, counter.Value())
}

func TestSumHandlerWithServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SumHandler))
	defer server.Close()

	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{"Simple", []int{1, 2, 3}, 6},
		{"Empty", []int{}, 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := http.Post(server.URL, "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Status code = %v, want %v", resp.StatusCode, http.StatusOK)
			}

			var result int
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				t.Fatal(err)
			}

			if result != tt.expected {
				t.Errorf("Result = %v, want %v", result, tt.expected)
			}
		})
	}
}
