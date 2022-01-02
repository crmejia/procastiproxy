package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// what am I testing??? test domains are inserted correctly
// by testing valid and invalid gets on the already tested go hashmap?
func TestParseBlockList(t *testing.T) {
	input := []string{"google.com", "example.net"}
	parseBlockList(&input)
	for _, v := range input {
		if !blockList[v] {
			t.Errorf("%s was not inserted in the blocklist", v)
		}
	}
}

func TestParseOfficeHours(t *testing.T) {
	test := []struct {
		name      string
		startTime string
		endTime   string
		err       error
	}{
		{
			name:      "test correct input no errors",
			startTime: "10:00",
			endTime:   "12:00",
			err:       nil,
		},
		{
			name:      "test noTimeProvidedError",
			startTime: "",
			endTime:   "12:00",
			err:       noTimeProvidedError,
		},
		{
			name:      "test malformedInputError (extra :00)",
			startTime: "12:00:11",
			endTime:   "12:00",
			err:       malformedInputError,
		},
		{
			name:      "test malformedInputError (gibberish)",
			startTime: "gibberish",
			endTime:   "12:00",
			err:       malformedInputError,
		},
	}
	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			err := parseOfficeHours(tc.startTime, tc.endTime)
			if err != tc.err {
				t.Errorf("expected no errors got: %s", err.Error())
			}
		})
	}
}

func TestProxy(t *testing.T) {
	test := []struct {
		name               string
		method             string
		target             string
		block              string
		statusCode         int
		officeHoursEnabled bool
		startTime          time.Time
		endTime            time.Time
	}{
		{
			name:       "test unblocked request",
			method:     http.MethodGet,
			target:     "http://example.com",
			block:      "",
			statusCode: http.StatusOK,
		},
		{
			name:       "test blocked request",
			method:     http.MethodGet,
			target:     "http://example.com",
			block:      "example.com",
			statusCode: http.StatusForbidden,
		},
		{
			name:               "test request is not blocked outside office hours ",
			method:             http.MethodGet,
			target:             "http://example.com",
			block:              "example.com",
			statusCode:         http.StatusOK,
			officeHoursEnabled: true,
			startTime:          time.Now().Add(time.Hour),
			endTime:            time.Now().Add(time.Hour * 3),
		},
		{
			name:               "test request is blocked during office hours",
			method:             http.MethodGet,
			target:             "http://example.com",
			block:              "example.com",
			statusCode:         http.StatusForbidden,
			officeHoursEnabled: true,
			startTime:          time.Now().Add(time.Hour * -1), //time - 1
			endTime:            time.Now().Add(time.Hour * 3),
		},
	}
	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			blockList = map[string]bool{
				tc.block: true,
			}
			officeHoursEnabled = tc.officeHoursEnabled
			startTime = tc.startTime
			endTime = tc.endTime

			request := httptest.NewRequest(tc.method, tc.target, nil)
			responseRecorder := httptest.NewRecorder()
			proxy(responseRecorder, request)
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("want %d, got %d", tc.statusCode, responseRecorder.Code)
			}
		})
	}
}

//test the program exits on empty blocklist
// once gain testing lang this? https://stackoverflow.com/questions/26225513/how-to-test-os-exit-scenarios-in-go/45379980

func TestAdminAddRemoveDomain(t *testing.T) {
	test := []struct {
		name       string
		method     string
		handler    http.HandlerFunc
		targetURL  string
		blocklist  map[string]bool
		key        string
		statusCode int
		want       bool
	}{
		{
			name:       "test blocking site",
			method:     http.MethodGet,
			handler:    adminBlockHandler,
			targetURL:  "http://localhost:8080/admin/block/example.net",
			blocklist:  map[string]bool{},
			key:        "example.net",
			statusCode: http.StatusOK,
			want:       true,
		},
		{
			name:       "test removing site",
			method:     http.MethodGet,
			handler:    adminUnblockHandler,
			targetURL:  "http://localhost:8080/admin/block/example.net",
			blocklist:  map[string]bool{"example.net": true},
			key:        "example.net",
			statusCode: http.StatusOK,
			want:       false,
		},
	}
	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			blockList = tc.blocklist
			//check that blocklist is in the correct state
			got := blockList[tc.key]
			if tc.want == got {
				//could be written better
				t.Errorf("want %t, got %t", tc.want, got)
			}

			request := httptest.NewRequest(tc.method, tc.targetURL, nil)

			responseRecorder := httptest.NewRecorder()
			tc.handler(responseRecorder, request)

			// check the response is correct, after all this is still a server
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("want %d, got %d", tc.statusCode, responseRecorder.Code)
			}

			//check that the hash value is changed.
			got = blockList[tc.key]
			if tc.want != got {
				//could be written better
				t.Errorf("want %t, got %t", tc.want, got)
			}
		})
	}
}
