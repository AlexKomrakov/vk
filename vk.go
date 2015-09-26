package main

type User struct {
	Id          int    `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Sex         int    `json:"sex"`
	Status      string `json:"status"`
	Bdate       string `json:"bdate"`
	Deactivated string `json:"deactivated"`
	City        City   `json:"city"`
	MobilePhone string `json:"mobile_phone"`
	HomePhone   string `json:"home_phone"`
}

type City struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}