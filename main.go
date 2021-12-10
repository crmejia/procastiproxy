package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var blockList map[string]bool = map[string]bool{
	"google.com":     true,
	"www.google.com": true,
}

func proxy(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Hostname())
	if _, ok := blockList[r.Host]; ok {
		// should i be creating a response object?
		//resp := http.Response{Request: r}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		//https://stackoverflow.com/questions/37863374/whats-the-difference-between-responsewriter-write-and-io-writestring
		jsonEncoder := json.NewEncoder(w)
		jsonEncoder.Encode(`{"message": "forbidden"}`)
	} else {
		// TODO do other methods
		resp, err := http.Get(r.RequestURI)

		if err != nil {
			fmt.Fprint(w, err)
		} else { // using else to avoid having naked return at the end
			// TODO what does resp.Write(w) mean?
			// it means you take your get response and send(write) it down the pipe
			resp.Write(w)
		}
	}
}
func main() {
	http.HandleFunc("/", proxy)
	http.ListenAndServe(":8080", nil)
}
