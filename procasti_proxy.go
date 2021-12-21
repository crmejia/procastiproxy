package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var blockList = map[string]bool{}
var startTime time.Time
var endTime time.Time
var officeHoursEnabled bool

const (
	startHour = 8
	startMin  = 0
	endHour   = 10
	endMin    = 0
)

func parseTime(hour, min int) time.Time {
	clock := time.Now()
	t := time.Date(clock.Year(), clock.Month(), clock.Day(), hour, min, clock.Second(), clock.Nanosecond(), clock.Location())
	return t
}

//Should this be a method? would it make it more "natural"
func proxy(w http.ResponseWriter, r *http.Request) {
	inHours := true
	if officeHoursEnabled && (startTime.After(time.Now()) || endTime.Before(time.Now())) {
		inHours = false
	}

	if blockList[r.URL.Hostname()] && inHours {
		// should I be creating a response object?
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
		entryUrl, err := url.Parse(entry)
		if err != nil {
			return err
		}
		blockList[entryUrl.Hostname()] = true
	}
	return nil
}

var mutex sync.Mutex

func adminHandler(w http.ResponseWriter, r *http.Request, b bool) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) == 4 {
		mutex.Lock()
		blockList[path[3]] = b
		mutex.Unlock()
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

	startTime = parseTime(startHour, startMin)
	endTime = parseTime(endHour, endMin)

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
