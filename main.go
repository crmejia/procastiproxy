package main

import (
	"fmt"
	"net/http"
)

var blockList map[string]bool = map[string]bool{
	"http://www.google.com/": true,
}

func proxy(w http.ResponseWriter, r *http.Request) {
	// TODO determine how the response is supposed to work
	if _, ok := blockList[r.RequestURI]; ok {
		fmt.Println(blockList, r.RequestURI)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "forbidden")
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
