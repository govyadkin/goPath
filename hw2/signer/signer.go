package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}

	in := make(chan interface{})
	var out chan interface{}

	for _, item := range jobs {
		out = make(chan interface{})
		wg.Add(1)
		go func(j job, in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			j(in, out)
			close(out)
		}(item, in, out, wg)
		in = out
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	outerWg := &sync.WaitGroup{}
	dataSignerMd5Mutex := &sync.Mutex{}
	for data := range in {
		outerWg.Add(1)
		go func(data interface{}) {
			defer outerWg.Done()
			convertedData := strconv.Itoa(data.(int))
			var result1, result2 string
			innerWg := &sync.WaitGroup{}
			innerWg.Add(2)
			go func() {
				defer innerWg.Done()
				result1 = DataSignerCrc32(convertedData)
			}()
			go func() {
				defer innerWg.Done()
				dataSignerMd5Mutex.Lock()
				md5 := DataSignerMd5(convertedData)
				dataSignerMd5Mutex.Unlock()
				result2 = DataSignerCrc32(md5)
			}()
			innerWg.Wait()
			out <- result1 + "~" + result2
		}(data)
	}
	outerWg.Wait()
}

func MultiHash(in, out chan interface{}) {
	outerWg := &sync.WaitGroup{}

	for data := range in {

		outerWg.Add(1)

		go func(data interface{}) {
			defer outerWg.Done()

			result := make([]string, 6)
			convertedData := data.(string)

			mutex := &sync.Mutex{}
			innerWg := &sync.WaitGroup{}

			for th := 0; th < 6; th++ {
				innerWg.Add(1)
				go func(th int) {
					defer innerWg.Done()
					hash := DataSignerCrc32(strconv.Itoa(th) + convertedData)
					mutex.Lock()
					result[th] = hash
					mutex.Unlock()
				}(th)
			}

			innerWg.Wait()

			mutex.Lock()
			out <- strings.Join(result, "")
			mutex.Unlock()
		}(data)
	}

	outerWg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for data := range in {
		results = append(results, data.(string))
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
}
