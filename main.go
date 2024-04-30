package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"context"
)

type Factorials struct {
	A int `json:"a"`
	B int `json:"b"`
}

func checkInput(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var input Factorials
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Incorrect input", http.StatusBadRequest)
			return
		}
		if input.A < 0 || input.B < 0 {
			http.Error(w, "Incorrect input", http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "input", input)
		next(w, r.WithContext(ctx), ps)
	}
}

func calculateFactorial(n int) int {
	if n == 0 {
		return 1
	}
	result := 1
	for i := 1; i <= n; i++ {
		result *= i
	}
	return result
}

func calculateFactorialAsync(n int, resultChan chan<- int) {
	resultChan <- calculateFactorial(n)
}

func postFactorials(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	input := r.Context().Value("input").(Factorials)

	resultChan := make(chan int, 2)

	go calculateFactorialAsync(input.A, resultChan)
	go calculateFactorialAsync(input.B, resultChan)

	input.B = <-resultChan
	input.A = <-resultChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(input)
}

func main() {
	router := httprouter.New()
	router.POST("/calculate", checkInput(postFactorials))

	fmt.Println("Server running on: 8989")
	http.ListenAndServe(":8989", router)
}
