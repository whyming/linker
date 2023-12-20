package main

import (
	"fmt"
	"net/http"
)

func main() {

	url := "http://tb-server-rd-wuhaiming.bcc-bdbl.baidu.com:8765/zebrago/zs/template/ugcAddPoi/"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("\n\nerror: %v\n\n", err)
		return
	}
	defer res.Body.Close()

	for {
		buf := make([]byte, 500)
		n, err := res.Body.Read(buf)
		if err != nil {
			fmt.Printf("\n\nerr: %v\n\n", err)
			return
		}
		fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxx")
		fmt.Println(string(buf[:n]))
		fmt.Println("------------------------")
	}
}
