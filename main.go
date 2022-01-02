package main

import "log"

func main() {
	proxy, err := NewProxy()
	if err != nil {
		log.Fatal(err)
	}
	proxy.Run()
}
