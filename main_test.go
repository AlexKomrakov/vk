package main

import (
    "time"
    "fmt"
    "testing"
    vegeta "github.com/tsenart/vegeta/lib"
	"regexp"
	"strconv"
)


func TestMain(t *testing.T) {
    rate := uint64(1000) // per second
    duration := 30 * time.Second
    targeter := vegeta.NewStaticTargeter(vegeta.Target{
        Method: "GET",
        URL:    "http://api.vk.com/method/users.get?user_id=1&v=5.37",
    })
    attacker := vegeta.NewAttacker(vegeta.Timeout(5 * time.Second))

    var results vegeta.Results
    for res := range attacker.Attack(targeter, rate, duration) {
        results = append(results, res)
    }

    metrics := vegeta.NewMetrics(results)
    fmt.Println(metrics)
    fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}

func TestParallelism (t *testing.T) {
	start        := time.Now()
	user_id      := "1"
	start_offset := "0"

	res, _ := GetFollowersSimple(user_id, start_offset)
	re := regexp.MustCompile(`"count":(\d+),`)
	count, _ := strconv.Atoi(re.FindAllStringSubmatch(res, 1)[0][1])

	fmt.Println(count)
	fmt.Println("Starting Go Routines")

	ch := make(chan string)
	for i := 1;  i<=count / 1000; i++ {

		time.Sleep(1 * time.Millisecond)
		go func(user_id string, i int) {
			for z := 1;  z<= 50; z++ {
				res, err := GetFollowersSimple(user_id, strconv.Itoa(i * 1000))
				if (err == nil) {
					ch <- res
					return
				}
				time.Sleep(1 * time.Millisecond)
			}
			return
		}(user_id, i)
	}

	fmt.Println("Waiting To Finish")
	cnt := 0
	Loop:
		for {
			select {
			case  <-ch:
				cnt++
				fmt.Println(cnt)
			case <-time.After(30 * time.Second):
				break Loop
			}
		}

	fmt.Println("\nTerminating Program")

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}
