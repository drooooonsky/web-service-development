package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

// сюда писать код

func SingleHash(in <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for input := range in {
		input_str := strconv.Itoa(input)
		result := DataSignerCrc32(input_str) + "~" + DataSignerCrc32(DataSignerMd5(input_str))
		fmt.Println("!!!! SingleHash result", result)
		runtime.Gosched()
	}
}

func MultiHash(d string) string {
	var result string
	for th := 0; th <= 5; th++ {
		result += DataSignerCrc32(strconv.Itoa(th) + d)
	}
	return result
}

func main() {
	// workerInput := make(chan, int)

	data := []int{110, 111, 112, 113, 114, 115, 116, 117}
	fmt.Println(data)
	// ch1 := data[1]
	// fmt.Printf("ch1 %v\n", ch1)
	// ch2 := SingleHash(ch1)
	// fmt.Printf("ch2 %v\n", ch2)
	// ch3 := MultiHash(ch2)
	// fmt.Printf("ch3 %v\n", ch3)

	wg := &sync.WaitGroup{}
	Ch1 := make(chan int)
	for _, x := range data {
		wg.Add(1)
		go SingleHash(Ch1, wg)
		Ch1 <- x
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
