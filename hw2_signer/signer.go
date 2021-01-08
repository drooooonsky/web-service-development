package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, jobFunc := range jobs {
		wg.Add(1)
		out := make(chan interface{})

		go func(wg *sync.WaitGroup, jobFunc job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			jobFunc(in, out)
		}(wg, jobFunc, in, out)

		in = out
	}
	wg.Wait()

}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in {
		data := fmt.Sprintf("%v", input) //strconv.Itoa(input)
		crcMd5 := DataSignerMd5(data)
		wg.Add(1)
		go workSingleHash(wg, data, crcMd5, out)
	}
	wg.Wait()
}

func workSingleHash(wg *sync.WaitGroup, data string, crcMd5 string, out chan interface{}) {
	defer wg.Done()

	crc32Chan := make(chan string)
	crcMd5Chan := make(chan string)

	go func(ch chan string, data string) {
		res := DataSignerCrc32(data)
		ch <- res
	}(crc32Chan, data)

	go func(ch chan string, data string) {
		res := DataSignerCrc32(data)
		ch <- res
	}(crcMd5Chan, crcMd5)

	crc32Hash := <-crc32Chan
	crc32Md5Hash := <-crcMd5Chan

	out <- crc32Hash + "~" + crc32Md5Hash

}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		go workMultiHash(wg, i, out)
	}

	wg.Wait()
}

func workMultiHash(wg *sync.WaitGroup, h interface{}, ch chan interface{}) {

	wgInternal := &sync.WaitGroup{}
	hashArray := make([]string, 6)

	defer wg.Done()

	for th := 0; th < 6; th++ {
		wgInternal.Add(1)
		data := strconv.Itoa(th) + fmt.Sprintf("%v", h)
		go calculateMultiHash(wgInternal, data, hashArray, th)
	}
	wgInternal.Wait()
	multiHash := strings.Join(hashArray, "")

	ch <- multiHash
}

func calculateMultiHash(wg *sync.WaitGroup, s string, array []string, index int) {
	defer wg.Done()
	crc32hash := DataSignerCrc32(s)
	array[index] = crc32hash
}

func CombineResults(in, out chan interface{}) {
	var hashArray []string

	for i := range in {
		hashArray = append(hashArray, i.(string))
	}

	sort.Strings(hashArray)
	combineResults := strings.Join(hashArray, "_")
	out <- combineResults
}
