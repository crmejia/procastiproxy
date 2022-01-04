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

//var noTimeProvidedError = errors.New("office hour time not provided")
var malformedInputError = errors.New("input is malformed")

type Proxy struct {
	server             *http.Server
	mutex              sync.Mutex
	blockList          map[string]bool
	startTime          time.Time
	endTime            time.Time
	officeHoursEnabled bool
}

func NewProxy() (*Proxy, error) {
	proxy := Proxy{}
	proxy.blockList = make(map[string]bool, 1)
	err := proxy.parseArgs()

	return &proxy, err
}

func (p *Proxy) Run() {
	r := http.NewServeMux()
	r.HandleFunc("/admin/block/", p.adminBlockHandler)
	r.HandleFunc("/admin/unblock/", p.adminUnblockHandler)
	r.HandleFunc("/", p.proxyHandler)
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
func (p *Proxy) proxyHandler(w http.ResponseWriter, r *http.Request) {
	inHours := true
	if p.officeHoursEnabled && (p.startTime.After(time.Now()) || p.endTime.Before(time.Now())) {
		inHours = false
	}

	if p.blockList[r.URL.Hostname()] && inHours {
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

func (p *Proxy) adminHandler(w http.ResponseWriter, r *http.Request, b bool) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) == 4 {
		p.mutex.Lock()
		p.blockList[path[3]] = b
		p.mutex.Unlock()
	} else {
		fmt.Fprintf(w, "malformed path %s", r.URL.Path)
	}
}

//TODO can these methods be simplified in a away that there's a single handler for both actions???
func (p *Proxy) adminBlockHandler(w http.ResponseWriter, r *http.Request) {
	p.adminHandler(w, r, true)
}
func (p *Proxy) adminUnblockHandler(w http.ResponseWriter, r *http.Request) {
	p.adminHandler(w, r, false)
}
func (p *Proxy) parseArgs() error {
	bl := pflag.StringSlice("blocklist", nil, "comma-separated list of hostnames to block")
	officeHourStartTime := pflag.String("starttime", "", "Office Hour Start time HH:MM")
	officeHourEndTime := pflag.String("endtime", "", "Office Hour Start time HH:MM")
	pflag.Parse()
	if err := p.parseOfficeHours(*officeHourStartTime, *officeHourEndTime); err != nil {
		return err
	}
	if err := p.parseBlockList(bl); err != nil {
		return err
	}
	return nil
}

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

func (p *Proxy) parseOfficeHours(officeHourStartTime string, officeHourEndTime string) error {
	if officeHourStartTime == "" || officeHourEndTime == "" {
		//if no time was provided, disable office hours
		p.officeHoursEnabled = false
		return nil
	}
	var err error
	p.startTime, err = parseTime(officeHourStartTime)
	if err != nil {
		return err
	}
	p.endTime, err = parseTime(officeHourEndTime)
	if err != nil {
		return err
	}
	p.officeHoursEnabled = true
	return nil
}

func (p *Proxy) parseBlockList(list *[]string) error {
	// its   to have a *[]string but I guess the point of this function is to remove the "weirdness" by parsing
	if *list == nil {
		return errors.New("no block list provided. ProcastiProxy won't run")
	}
	for i := 0; i < len((*list)[0:]); i++ {
		entry := (*list)[i]
		entryUrl, err := url.Parse(entry)
		if err != nil {
			return err
		}
		p.blockList[entryUrl.Hostname()] = true
	}
	return nil
}
