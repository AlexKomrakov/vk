package main

import (
	"net/http"
	"fmt"
	"time"
	"log"
	"io/ioutil"
//    "strings"
	"math/rand"
	//"net"
	"strconv"
	"net/url"
	//"crypto/tls"
	"encoding/json"
	"errors"
)

var (
	TOKENS = []string{""}
	WORDS  = []string{"увещевательный", "подличать", "скатерка", "пропихивать", "сыродельня", "отпасти", "дымоотводный"}
)

type GetMembersStruct struct {
	Response GetMembersResponse `json:"response"`
}

type GetMembersResponse struct {
	Items []User   `json:"items"`
	ItemsCount int `json:"items_count"`
	TotalCount int `json:"total_count"`
}

func GetToken() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return TOKENS[rand.Intn(len(TOKENS))]
}

type GroupRequest struct {
	Name   string
	Offset string
}

func GroupsGetMembers(req GroupRequest) (result GetMembersStruct, err error)  {
	u      := "https://api.vk.com/method/execute"
	token  := GetToken()
	count  := "1000"
	v      := "5.37"
	fields := "sex, status, contacts, city, bdate"
	code  := `
        var count = `+ count +`, v = "` + v + `", offset = ` + req.Offset+ `, group_id = "` + req.Name + `", fields = "` + fields + `", iteration = 1, result;
        var res = API.groups.getMembers({
                "v": v,
                "count": count,
                "group_id": group_id,
                "fields": fields,
                "offset": offset
            });
		result = {
			"items": res.items,
			"items_count": res.items.length,
			"total_count": res.count,
		};
		while(result.total_count > result.items_count + offset && iteration < 25) {
			iteration = iteration + 1;
            var res = API.groups.getMembers({
                "v": v,
                "count": count,
                "offset": offset + result.items_count,
                "group_id": group_id,
				"fields": fields
            });
            result.items       = result.items + res.items;
            result.items_count = result.items.length;
			result.total_count = res.count;
        }
        return result;
        `

	query := url.Values{"access_token" : []string{token}, "code" : []string{code}}
	res, err := http.Get(u + "?" + query.Encode())
	if err != nil {
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return
	}

	if result.Response.ItemsCount == 0 {
		return result, errors.New("Empty result")
	}

	return
}



type Handler struct {}

func NewsfeedSearch(q string) {
	u     := "https://api.vk.com/method/execute"
	code  := `
        var count = 200, v = "5.37", iteration = 1, result;
        var rq = API.newsfeed.search({
                "q": "` + q + `",
                "v": v,
                "count": count
            });
		result = {
			"items": rq.items,
			"items_count": rq.items.length,
			"total_count": rq.total_count,
			"next_from": rq.next_from
		};
		while(result.total_count > result.items_count && iteration <= 10) {
			iteration = iteration + 1;
            var rq = API.newsfeed.search({
                "q": "` + q + `",
                "v": v,
                "count": count,
                "start_from": result.next_from
            });
            result.items = result.items + rq.items;
            result.items_count = result.items.length;
            result.next_from = rq.next_from;
        }
        result.items = [];
        return result;
        `

	token := GetToken()

	qq := url.Values{}
	qq.Set("access_token", token)
	qq.Set("code", code)

	res, err := http.Get(u + "?" + qq.Encode())
	if err != nil {
		log.Print(err)
	}
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Print(err)
	}
	fmt.Print(string(response))
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	uri := r.URL.Path
	q := r.URL.Query()
	token := GetToken()
	q.Set("access_token", token)
//	word := words[rand.Intn(len(words))]
	user_id := strconv.Itoa(rand.Intn(6000000))
	offset := "0"
	code := `
        var users = []; var offset = ` + offset + `; var start_offset = ` + offset + `; var count = 1000; var iteration = 1; var totalUsers = 0;
        var rq = API.users.getFollowers({
                "user_id": "` + user_id + `",
                "v": "5.28", "count": count, "offset": offset
            });
        offset=offset+count;
        users = users + rq.items;
        var accounted = rq.items.length;
        totalUsers = rq.count;
        while(totalUsers > 0 && totalUsers > accounted && iteration <= 24){
            rq = API.users.getFollowers({
                "user_id": "` + user_id + `",
                "v": "5.37", "count": count, "offset": offset
            });
            users = users + rq.items;
            offset=offset+count;
            accounted = accounted + rq.items.length;
            iteration = iteration + 1;
            totalUsers = rq.count;
        }
        if(parseInt(totalUsers)==0 && totalUsers+"" == ""){
            return {
                "error": {
                    "error_code": 0
                }
            };
        }
        return {
            "users": users,
            "total": totalUsers,
            "offset": start_offset,
            "count": users.length,
        };
        `
	q.Set("code", code)
//	q.Set("user_id", user_id)
	http.Get("https://api.vk.com/method/users.get?v=5.37&user_id=1")
//	res, err := http.Get("https://api.vk.com/" + uri + "?" + q.Encode())
//	if err != nil {
//		log.Print(err)
//	}
//	response, err := ioutil.ReadAll(res.Body)
//	res.Body.Close()
//	if err != nil {
//		log.Print(err)
//	}
//	if (strings.Contains(string(response), "users")) {
		w.WriteHeader(http.StatusOK)
//	} else {
//		w.WriteHeader(http.StatusNotFound)
//	}
	fmt.Fprint(w, q.Encode())
	fmt.Fprint(w, uri)
//	fmt.Fprint(w, string(response))
	return
}
func GetFollowers(user_id, offset string) string {
	token := GetToken()
	code := `
        var users = []; var offset = ` + offset + `; var start_offset = ` + offset + `; var count = 1000; var iteration = 1; var totalUsers = 0;
        var rq = API.users.getFollowers({
                "user_id": "` + user_id + `",
                "v": "5.28", "count": count, "offset": offset
            });
        offset=offset+count;
        users = users + rq.items;
        var accounted = rq.items.length;
        totalUsers = rq.count;
        while(totalUsers > 0 && totalUsers > accounted && iteration <= 24){
            rq = API.users.getFollowers({
                "user_id": "` + user_id + `",
                "v": "5.37", "count": count, "offset": offset
            });
            users = users + rq.items;
            offset=offset+count;
            accounted = accounted + rq.items.length;
            iteration = iteration + 1;
            totalUsers = rq.count;
        }
        if(parseInt(totalUsers)==0 && totalUsers+"" == ""){
            return {
                "error": {
                    "error_code": 0
                }
            };
        }
        return {
            "users": users,
            "total": totalUsers,
            "offset": start_offset,
            "count": users.length,
        };
        `

	baseUrl, err := url.Parse("https://api.vk.com/method/execute")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("code", code)
	params.Add("access_token", token)

	baseUrl.RawQuery = params.Encode()
	fmt.Println(baseUrl)

	res, err := http.Get(baseUrl.String())

	if err != nil {
		log.Print(err)
	}
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Print(err)
	}

	return string(response)
}

func GetFollowersSimple(user_id, offset string) (resp string, err error) {
//	timeout := time.Duration(5 * time.Second)
//	dialer := &net.Dialer{
//		LocalAddr: &net.TCPAddr{IP: net.IPAddr{IP: net.IPv4zero}.IP, Zone: net.IPAddr{IP: net.IPv4zero}.Zone},
//		KeepAlive: 30 * time.Second,
//		Timeout:   timeout,
//	}
//	client := http.Client{
//		Transport: &http.Transport{
//			Proxy: http.ProxyFromEnvironment,
//			Dial: dialer.Dial,
//			ResponseHeaderTimeout: timeout,
//			TLSHandshakeTimeout:   5 * time.Second,
//			MaxIdleConnsPerHost:   10000,
//			TLSClientConfig: &tls.Config{
//				InsecureSkipVerify: true,
//			},
//		},
//	}
	res, err := http.Get("https://api.vk.com/method/users.getFollowers?user_id=" +user_id+ "&v=5.37&access_token=4150078057cc11e9b316c5e60a0146147dad1c8b8e9486d18352f3eea57472a9eea676e308c69fc3aca85&count=1000" + "&offset=" + offset)
	if err != nil {
		//fmt.Println(err.Error())
		return
	}
	defer res.Body.Close()

	response, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	return string(response), nil
}

func main() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 500
	handler := new(Handler)
	s := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
