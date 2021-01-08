package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

func SingleHash(in <-chan int, out chan<- string, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	for input := range in {
		input_str := strconv.Itoa(input)
		mu.Lock()
		dataMd5 := DataSignerMd5(input_str)
		mu.Unlock()
		result := DataSignerCrc32(input_str) + "~" + DataSignerCrc32(dataMd5)
		fmt.Println("!!!! SingleHash result", result)
		out <- result
		runtime.Gosched()
	}
}

func MultiHash(in <-chan string, out chan<- string, wg *sync.WaitGroup) {
	var result string
	defer wg.Done()
	input := <-in
	for th := 0; th <= 5; th++ {
		result += DataSignerCrc32(strconv.Itoa(th) + input)
	}
	out <- result
	fmt.Println("!!!! MultiHash result", result)
}

func CombineResults(in <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	x := <-in
	fmt.Println(x)
}

func main() {
	data := []int{0, 111, 112, 113, 114, 115, 116, 117}
	fmt.Println(data)

	// WaitGroup чтобы дождаться выполнения корутин
	wg := &sync.WaitGroup{}
	// Mutex чтобы залочиться для функции DataSignerMd5
	mu := &sync.Mutex{}
	// канал для обмена между исходными данными и SingleHash
	Ch1 := make(chan int)
	// канал для обмена данными между SingleHash и MultiHash
	Ch2 := make(chan string)
	// канал для обмена данными между MultiHash и CombineResults
	Ch3 := make(chan string)

	for _, x := range data {
		wg.Add(1)
		go SingleHash(Ch1, Ch2, wg, mu)
		Ch1 <- x
		wg.Add(1)
		go MultiHash(Ch2, Ch3, wg)
		wg.Add(1)
		go CombineResults(Ch3, wg)
	}
	close(Ch1)

	wg.Wait()

	// fmt.Scanln()
	// time.Sleep(1000 * time.Millisecond)

	// LOOP:
	// 	for {
	// 		select {
	// 		case d, ok := <-Ch1:
	// 			fmt.Println("!!!", d, ok)
	// 		default:
	// 			fmt.Println("quit")
	// 			break LOOP
	// 		}
	// 	}

}
