package main

import (
	"fmt"
	"net/http"
	"steam/pkg/auth"
)

func main() {
	httpClient := http.Client{}
	mgr := auth.Core{}

	mgr.Init(
		&httpClient,
		auth.LoginInfo{
			UserName: "user",
			Password: "password",
		})
	mgr.SetHttpParam(5000, "http://127.0.0.1:1234")

	err := mgr.Login()
	if err != nil {
		fmt.Println(err.Error())
	}
}
