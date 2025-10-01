package easter

import "strings"

const (
	asciiArt = `      .-""""-.
    .'  _  _  '.
   /   (o)(o)   \
  |      ^^      |
  |  you found   |
  |     me!      |
   \  (______ ) /
    '.        .'
      '-.__.-'`
	asciiEgg = "```text\n" + asciiArt + "\n```"
)

// Message returns the Easter egg ASCII art when the input contains the word
// "easter" (case-insensitive). The boolean indicates whether the Easter egg
// should be used for the response.
func Message(input string) (string, bool) {
	if strings.Contains(strings.ToLower(input), "easter") {
		return asciiEgg, true
	}
	return "", false
}
