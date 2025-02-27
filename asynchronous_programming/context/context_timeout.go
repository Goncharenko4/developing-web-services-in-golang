package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func worker(ctx context.Context, workerNum int, out chan<- int) {
	waitTime := time.Duration(rand.Intn(100)+10) * time.Millisecond
	fmt.Println(workerNum, "студент ", waitTime)
	select {
	case <-ctx.Done():
		return
	case <-time.After(waitTime):
		fmt.Println("студент", workerNum, "придумал вопрос")
		out <- workerNum
	}
}

func main() {
	workTime := 50 * time.Millisecond
	// WithTimeout - это тип контекста, который может завершаться по какому-то таймауту
	ctx, _ := context.WithTimeout(context.Background(), workTime) // context - неизменяемая структура данных, передавать первым аргументом
	result := make(chan int, 1)

	for i := 0; i <= 10; i++ {
		go worker(ctx, i, result) // после завершения воркера контекст рассылает Done()
	}

	totalFound := 0
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case foundBy := <-result:
			totalFound++
			fmt.Println("студент", foundBy, "задал вопрос")
		}
	}
	fmt.Println("всего вопросов", totalFound)
	time.Sleep(time.Second)
}
