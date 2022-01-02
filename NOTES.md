#NOTES
## Improvements
1. Change blocklist to `map[url.net]bool` to leverage URL parsing.
2. Office hours is set to the current day. It shold work on schedule.
3. 

##observations
* `url.Parse(entry)` is quite strict on parsing. Had to not include the www in the input.
* In general parsing and edge cases are not completely covered as they're not the goal of the exercise.