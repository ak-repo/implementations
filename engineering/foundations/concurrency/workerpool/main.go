package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func worker(ctx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup) {

	for {
		select {
		case <-ctx.Done():
			// cancellation signal received
			fmt.Printf("Worker %d stopping\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				// channel closed
				return
			}

			fmt.Printf("Worder %d started %d job\n", id, job)
			time.Sleep(time.Second)
			fmt.Printf("Worder %d finidhed %djob\n", id, job)

			wg.Done()
		}

	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobs := make(chan int)
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		go worker(ctx, i, jobs, &wg)
	}

	for j := 1; j <= 5; j++ {
		wg.Add(1)
		jobs <- j
	}
	close(jobs)

	// wait or cancel
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
		return
	case <-done:
		fmt.Fprintln(w, "All jobs completed")
	}
}

func main() {
	http.HandleFunc("/process", handler)
	http.ListenAndServe(":8080", nil)

}
