package main

import (
    "time"
    "fmt"
    "testing"
	"regexp"
	"strconv"
	"gopkg.in/mgo.v2"
)

var session *mgo.Session

func Init() {
	sess, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	session = sess
}

func grabber(c chan int, ch_res chan string, err_ch chan int) {
	fmt.Println("Starting grabber")
	for i := 0; i < 1000; i++ {
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
	errors, cnt, count_st := 0, 0, 0
	Loop:
		for {
			select {
			case  <-res_ch:
				cnt++
				count_st++
//				fmt.Println(cnt)
				if cnt == count / 1000 {
					break Loop
				}
			case <-err_ch:
				errors++
			case <-ticker.C:
				fmt.Printf("Всего %d / Ошибок/сек %d (%d записей/сек) \n", cnt, errors, count_st/10)
				count_st = 0
				errors = 0
			case <-time.After(30 * time.Second):
				break Loop
			}
		}

	fmt.Println("\nTerminating Program")

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}

func GMGrabbers(c chan GroupRequest, ch_res chan GetMembersStruct, err_ch chan int) {
	fmt.Println("Starting grabbers")
	for i := 0; i < 1000; i++ {
		go func() {
			for {
				select {
				case request:= <-c:
					res, err := GroupsGetMembers(request)
					if err != nil {
						c <- request
						err_ch <- 1
					} else {
						ch_res <- res
					}
				}
			}
		} ()
	}
}

func TestGetGroupMembers(t *testing.T) {
	Init()

	results_ch := make(chan GetMembersStruct, 200)
	request_ch := make(chan GroupRequest, 500)
	errors_ch  := make(chan int, 10)

	GMGrabbers(request_ch, results_ch, errors_ch)

	group_name := "strategicmusic"

	fmt.Println("Initial request")
	result, err := GroupsGetMembers(GroupRequest{group_name, "0"})
	if err != nil {
		panic(err)
	}
	results_ch <- result

	sent := 1
	fmt.Println(result.Response.TotalCount)
	for offset := 25000; offset < result.Response.TotalCount; offset += 25000 {
		sent++
		fmt.Println("Request with offset: " + strconv.Itoa(offset))
		request_ch <- GroupRequest{group_name, strconv.Itoa(offset)}
	}

	ticker := time.NewTicker(time.Duration(10) * time.Second)
	defer ticker.Stop()

	errors, cnt, count_st := 0, 0, 0
	Loop:
	for {
		select {
		case result := <-results_ch:
			cnt++
			count_st++
			_, err = SaveInDB("vk", group_name, result.Response.Items)
			if err != nil {
				panic(err)
			}

			if (cnt == sent) {
				fmt.Println("Finish")
				break Loop
			}
		case <-errors_ch:
			errors++
		case <-ticker.C:
			fmt.Printf("Всего %d / Ошибок/сек %d (%d записей/сек) \n", cnt, errors, count_st / 10)
			count_st = 0
			errors = 0
		}
	}

	fmt.Println("Done")
}

func SaveInDB(db, col string, data []User) (*mgo.BulkResult, error) {
	b := session.DB(db).C(col).Bulk()
	for _, val := range data {
		b.Insert(val)
	}

	return b.Run()
}