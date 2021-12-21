package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var blockList = map[string]bool{}

//Should this be a method? would it make it more "natural"
func proxy(w http.ResponseWriter, r *http.Request) {
	if blockList[r.URL.Hostname()] {
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

func parseBlockList(list *[]string) error {
	// its   to have a *[]string but I guess the point of this function is to remove the "weirdness" by parsing
	for i := 0; i < len((*list)[0:]); i++ {
		entry := (*list)[i]
		url, err := url.Parse(entry)
		if err != nil {
			return err
		}
		blockList[url.Hostname()] = true
	}
	return nil
}

func adminHandler(w http.ResponseWriter, r *http.Request, b bool) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) == 4 {
		blockList[path[3]] = b
		fmt.Println(blockList)
	} else {
		fmt.Fprintf(w, "malformed path %s", r.URL.Path)
	}
}
func adminBlockHandler(w http.ResponseWriter, r *http.Request) {
	adminHandler(w, r, true)
}
func adminUnblockHandler(w http.ResponseWriter, r *http.Request) {
	adminHandler(w, r, false)
}

func run() {
	bl := pflag.StringSlice("blocklist", nil, "comma-separated list of hostnames to block")
	pflag.Parse()

	if *bl != nil {
		//insert into the dictionary
		parseBlockList(bl)

		http.HandleFunc("/admin/block/", adminBlockHandler)
		http.HandleFunc("/admin/unblock/", adminUnblockHandler)
		http.HandleFunc("/", proxy)
		http.ListenAndServe(":8080", nil)
	}

	log.Fatal("no block list provided. ProcastiProxy won't run")
}
