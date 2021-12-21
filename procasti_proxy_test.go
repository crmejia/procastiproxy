package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// what am I testing??? test domains are inserted correctly
// by testing valid and invalid gets on the already tested go hashmap?
func TestParseArgs(t *testing.T) {
	input := []string{"google.com", "example.net"}
	//var input []string
	parseBlockList(&input)

	for _, v := range input {
		//t.Logf(v)
		if !blockList[v] {
			t.Errorf("%s was not inserted in the blocklist", v)
		}
	}

}

func TestProxy(t *testing.T) {
	test := []struct {
		name       string
		method     string
		target     string
		block      string
		statusCode int
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
	}
	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.target, nil)
			blockList = map[string]bool{
				tc.block: true,
			}
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

//func TestAdminAddRemoveDomain(t *testing.T) {
//	test := []struct {
//		name    string
//		domain  string
//		allowed bool
//	}{
//		{
//			name:    "test adding site",
//			domain:  "example.net",
//			allowed: false,
//		},
//		{
//			name:    "test removing site",
//			domain:  "example.net",
//			allowed: true,
//		},
//	}
//
//}
