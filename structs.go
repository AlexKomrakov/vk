package main

type User struct {
	Id          int    `json:"id" bson:"_id"`
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

type GetMembersStruct struct {
	Response GetMembersResponse `json:"response"`
}

type GetUsersStruct struct {
	Response []User `json:"response"`
}

type GetMembersResponse struct {
	Items      []User `json:"items"`
	ItemsCount int    `json:"items_count"`
	TotalCount int    `json:"total_count"`
}

type GroupRequest struct {
	Name   string
	Offset string
}

type UsersRequest struct {
	Start  int
	Count  int
}