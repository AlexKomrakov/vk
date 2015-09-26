package main

import (
	"time"
	"fmt"
	"testing"
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

func UsersGrabbers(c chan UsersRequest, ch_res chan GetUsersStruct, err_ch chan int, workers int) {
	fmt.Println("Starting grabbers")
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case request:= <-c:
					res, err := UsersGet(request)
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

func GetAllUsers(workers int) {
	Init()

	results_ch := make(chan GetUsersStruct, 200)
	request_ch := make(chan UsersRequest, workers)
	errors_ch  := make(chan int, 500)

	UsersGrabbers(request_ch, results_ch, errors_ch, workers)

	sent := 0
	fmt.Println("Sending tasks")
	go func() {
		for i := 1; i <= 300000000; i = i + 100 {
			sent++
			request_ch <- UsersRequest{i, 100}
		}
	}()

	ticker := time.NewTicker(time.Duration(1) * time.Second)
	defer ticker.Stop()

	errors, cnt, count_st := 0, 0, 0
	Loop:
	for {
		select {
		case result := <-results_ch:
			cnt++
			count_st++
			_, err := SaveInDB("vk", "users", result.Response)
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
			fmt.Printf("Всего %d / Ошибок/сек %d (%d записей/сек) \n", cnt, errors, count_st / 1)
			count_st = 0
			errors = 0
		}
	}

	fmt.Println("Done")
}