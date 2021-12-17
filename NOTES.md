#NOTES
## Improvements
1. change blocklist to `map[url.net]bool` to leverage URL parsing.

##observations
`url.Parse(entry)` is quite strict on parsing. Had to not include the www in the input.