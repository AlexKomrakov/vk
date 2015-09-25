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
	ticker := time.NewTicker(time.Duration(10) * time.Second)
	defer ticker.Stop()

	start        := time.Now()
	user_id      := "1"
	start_offset := "0"

	res, _ := GetFollowersSimple(user_id, start_offset)
	re := regexp.MustCompile(`"count":(\d+),`)
	count, _ := strconv.Atoi(re.FindAllStringSubmatch(res, 1)[0][1])

	fmt.Println(count)
	fmt.Println("Starting Go Routines")

	ch     := make(chan int, 7000)
	res_ch := make(chan string)
	err_ch := make(chan int, 5000)

	go grabber(ch, res_ch, err_ch)

	for i := 1;  i<=count / 1000; i++ {
		val := i*1000
		ch <- val
	}

	fmt.Println("Waiting To Finish")
	errors, cnt, count := 0, 0, 0
	Loop:
		for {
			select {
			case  <-res_ch:
				cnt++
				count++
//				fmt.Println(cnt)
				if cnt == count / 1000 {
					break Loop
				}
			case <-err_ch:
				errors++
			case <-ticker.C:
				fmt.Printf("Всего %d / Ошибок/сек %d (%d записей/сек) \n", cnt, errors, count/10)
//				fmt.Println(cnt)
//				fmt.Println(count/10)
//				fmt.Println(errors)
				count = 0
				errors = 0
			case <-time.After(30 * time.Second):
				break Loop
			}
		}

	fmt.Println("\nTerminating Program")

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}

func grabber(c chan int, ch_res chan string, err_ch chan int) {
	fmt.Println("Starting grabber")
	for i := 0; i < 500; i++ {
		go func() {
			for {
				select {
				case val:= <-c:
					res, err := GetFollowersSimple("1", strconv.Itoa(val))
					if err != nil {
						c <- val
						err_ch <- 1
					} else {
						ch_res <- res
					}
				case <-time.After(5 * time.Second):
					fmt.Print("Break")
					break
				}
			}
		} ()
	}
}


func TestGo(t *testing.T) {
	ch := make(chan int) //создаем канал для передачи целых чисел (int)
	b:= <-ch   //читаем число из канала
	fmt.Println(b)
}
