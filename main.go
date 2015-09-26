package main

import (
	"net/http"
	"fmt"
	"time"
	"io/ioutil"
	"math/rand"
	"strconv"
	"net/url"
	"encoding/json"
	"errors"
)

var (
	TOKENS = []string{""}
)

func GetToken() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return TOKENS[rand.Intn(len(TOKENS))]
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

func UsersGet(req UsersRequest) (result GetUsersStruct, err error)  {
	u        := "https://api.vk.com/method/users.get"
	token    := GetToken()
	v        := "5.37"
	fields   := "sex, contacts, city, bdate"
	var user_ids string
	for i := req.Start; i < req.Start + req.Count; i++ {
		user_ids = user_ids + strconv.Itoa(i)
		if (i+1 != req.Start + req.Count) {
			user_ids = user_ids + ","
		}
	}

	query := url.Values{"access_token" : []string{token}, "v" : []string{v}, "fields" : []string{fields}}
	res, err := http.Get(u + "?" + query.Encode() + "&user_ids=" + user_ids)
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

	if len(result.Response) == 0 {
		return result, errors.New("Empty result")
	}

	return
}


type Handler struct {}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	uri := r.URL.Path
	q := r.URL.Query()
	token := GetToken()
	q.Set("access_token", token)
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
	http.Get("https://api.vk.com/method/users.get?v=5.37&user_id=1")

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, q.Encode())
	fmt.Fprint(w, uri)
	return
}

func StartServer() {
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

func main() {
	// StartServer()

	workers := 500
	GetAllUsers(workers)
}

