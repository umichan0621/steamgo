package main

import (
	"fmt"
	"steam/pkg/auth"
)

func main() {
	mgr := auth.Manager{}
	mgr.Init(auth.LoginInfo{
		UserName: "user",
		Password: "password",
	})
	mgr.SetHttpParam(5000, "http://127.0.0.1:1234")

	err := mgr.Login()
	if err != nil {
		fmt.Println(err.Error())
	}
}
