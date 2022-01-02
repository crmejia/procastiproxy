package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/pflag"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

//TODO make these vars
var blockList = map[string]bool{}
var startTime time.Time
var endTime time.Time
var officeHoursEnabled bool
var mutex sync.Mutex

var noTimeProvidedError = errors.New("office hour time not provided")
var malformedInputError = errors.New("input is malformed")

//func parseTime(hour, min int) time.Time {
func parseTime(inputTime string) (time.Time, error) {
	inputSlice := strings.Split(inputTime, ":")
	if len(inputSlice) == 2 {
		hour, err := strconv.Atoi(inputSlice[0])
		if err != nil {
			return time.Now(), err
		}
		min, err := strconv.Atoi(inputSlice[1])
		if err != nil {
			return time.Now(), err
		}
		clock := time.Now()
		return time.Date(clock.Year(), clock.Month(), clock.Day(), hour, min, clock.Second(), clock.Nanosecond(), clock.Location()), nil
	}
	return time.Now(), malformedInputError
}

func parseOfficeHours(officeHourStartTime string, officeHourEndTime string) error {
	if officeHourStartTime == "" || officeHourEndTime == "" {
		return noTimeProvidedError
	}
	var err error
	startTime, err = parseTime(officeHourStartTime)
	if err != nil {
		return err
	}
	endTime, err = parseTime(officeHourEndTime)
	if err != nil {
		return err
	}
	officeHoursEnabled = true
	return nil
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

type Proxy struct {
	mutex  sync.Mutex
	server *http.Server
}

func NewProxy() *Proxy {
	bl := pflag.StringSlice("blocklist", nil, "comma-separated list of hostnames to block")
	officeHourStartTime := pflag.String("starttime", "", "Office Hour Start time HH:MM")
	officeHourEndTime := pflag.String("endtime", "", "Office Hour Start time HH:MM")
	pflag.Parse()
	parseOfficeHours(*officeHourStartTime, *officeHourEndTime)

	if *bl != nil {
		//insert into the dictionary
		parseBlockList(bl)
	} else {
		log.Fatal("no block list provided. ProcastiProxy won't run")
	}
	proxy := Proxy{}
	return &proxy
}

func (p *Proxy) Run() {
	r := http.NewServeMux()
	r.HandleFunc("/admin/block/", adminBlockHandler)
	r.HandleFunc("/admin/unblock/", adminUnblockHandler)
	r.HandleFunc("/", proxyHandler)
	p.server = &http.Server{
		Handler: r,
		Addr:    "localhost:8080",
	}

	errs := make(chan error, 1)
	go func() {
		errs <- p.server.ListenAndServe()
	}()

	log.Fatal(<-errs)
}

//Should this be a method? would it make it more "natural"
func proxyHandler(w http.ResponseWriter, r *http.Request) {
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
		jsonEncoder.Encode(`{message: forbidden}`)
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
