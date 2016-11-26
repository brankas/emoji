// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	. "github.com/knq/emoji"
)

const (
	gemojiURL = "https://raw.githubusercontent.com/github/gemoji/master/db/emoji.json"
)

var (
	flagOut = flag.String("o", "gemoji_data.go", "out")
)

func main() {
	var err error

	flag.Parse()

	// generate data
	buf, err := generate()
	if err != nil {
		log.Fatal(err)
	}

	// write
	err = ioutil.WriteFile(*flagOut, buf, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

var replacer = strings.NewReplacer(
	"emoji.Gemoji", "Gemoji",
	"emoji.Emoji", "\n",
	"}}", "},\n}",
	", Description:", ", ",
	", Category:", ", ",
	", Aliases:", ", ",
	", Tags:", ", ",
	", UnicodeVersion:", ", ",
	", IOSVersion:", ", ",
)

var emojiRE = regexp.MustCompile(`\{Emoji:"([^"]*)"`)

func generate() ([]byte, error) {
	var err error

	// load gemoji data
	res, err := http.Get(gemojiURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// read all
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal
	var data Gemoji
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	// add header
	str := replacer.Replace(fmt.Sprintf(hdr, gemojiURL, data))

	// change the format of the unicode string
	str = emojiRE.ReplaceAllStringFunc(str, func(s string) string {
		var err error
		s, err = strconv.Unquote(s[len("{Emoji:"):])
		if err != nil {
			panic(err)
		}
		return "{" + strconv.QuoteToASCII(s)
	})

	// format
	return format.Source([]byte(str))
}

const hdr = `
package emoji

// GemojiData is the original Gemoji data.
//
// see: %s
var GemojiData = %#v
`
