package main

import (
	"fmt"
	"github.com/lukashes/db/client"
	"runtime"
	"sync"
	"time"
)

func main() {

	wg := sync.WaitGroup{}

	t := time.Now()

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			c, _ := client.New("http://localhost:8080")

			var s, d string

			s = fmt.Sprintf("value_%d", worker)
			key := fmt.Sprintf("key_of_worker_%d", worker)

			for i := 0; i < 100000; i++ {
				k := fmt.Sprintf("%s_%d", key, i)
				err := c.Add(k, s)
				if err != nil {
					panic(err)
				}
			}

			for i := 0; i < 100000; i++ {
				k := fmt.Sprintf("%s_%d", key, i)
				v, err := c.Get(k)
				if err != nil {
					panic(err)
				}
				v.Decode(&d)

				if d != s {
					panic(fmt.Errorf("Differet values %v != %v", s, d))
				}
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("Latency: %d ms\n", time.Now().Sub(t).Nanoseconds()/1000/1000)
}
