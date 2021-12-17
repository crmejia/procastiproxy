package main

import (
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

//test the program exits on empty blocklist
// once gain testing lang this? https://stackoverflow.com/questions/26225513/how-to-test-os-exit-scenarios-in-go/45379980
