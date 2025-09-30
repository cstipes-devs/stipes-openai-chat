package easter

import "strings"

const asciiEgg = `
      .-""""-.
    .'  _  _  '.
   /   (o)(o)   \
  |      ^^      |
  |    .----.    |
   \  (______ ) /
    '.        .'
      '-.__.-'
you found me!
`

// Message returns the Easter egg ASCII art when the input contains the word
// "easter" (case-insensitive). The boolean indicates whether the Easter egg
// should be used for the response.
func Message(input string) (string, bool) {
	if strings.Contains(strings.ToLower(input), "easter") {
		return asciiEgg, true
	}
	return "", false
}
