package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/eefret/gomdb"
	"github.com/odwrtw/yifysubs"
)

type movie struct {
	Title string
	Date  string
}

//TODO: check if scanned dir includes a video file, if not, dont add it to our list

// scan a folder for dirs and return them in a list
func scandirs(path string) []string {
	//r, _ := regexp.Compile(`.+\)`)
	count := 0
	dirs := 0
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Panicln("Reading path failed: ", err)
	}
	// this is ugly, find a better way so this doesn't loop dirs twice
	for _, f := range files {
		if f.IsDir() /*&& r.MatchString(f.Name()) && f.Name()[0] != '('*/ {
			dirs++
		}
	}
	founddirs := make([]string, dirs, dirs+1)
	for _, f := range files {
		if f.IsDir() /*&& r.MatchString(f.Name()) && f.Name()[0] != '('*/ {
			founddirs[count] = f.Name()
			count++
		}
	}
	return founddirs
}

// Get a raw dir title and extract information from it, should work for all YTS downloads
func getmovie(in string) (string, string) {
	rtitle, _ := regexp.Compile(`[^(]*`)
	ryear, _ := regexp.Compile(`\(.+?\)`)

	title := strings.TrimSpace(rtitle.FindString(in))
	date := strings.Trim(ryear.FindString(in), "()")

	return title, date
}

func getimdb(title string, date string) string {
	if title == "" || date == "" {
		return "NOTMOVIE"
	}
	query := &gomdb.QueryData{Title: title, Year: date}
	res, err := gomdb.MovieByTitle(query)
	if err != nil {
		//fmt.Println("Querying OMDB failed:", err)
	}
	if res.ImdbID == "" {
		return ""
	}
	return res.ImdbID
}

func main() {
	path := ""
	if len(os.Args) == 1 {
		path = "."
	} else {
		path = os.Args[1]
	}
	dirs := scandirs(path)
	m := new(movie)

	for _, dir := range dirs {
		m.Title, m.Date = getmovie(dir)
		filepath := path + `\` + dir + `\` // C:\target path\movie folder\
		if m.Title == "" || m.Date == "" {
			continue
		}
		imdb := getimdb(m.Title, m.Date)
		if imdb == "NOTMOVIE" || imdb == "" {
			continue
		}
		subs, err := yifysubs.GetSubtitles(imdb)

		if err != nil {
			//fmt.Println("Sub finding failed")
			continue
		}
		en := subs["english"][0]

		srt, err := os.Stat(filepath + m.Title + " subtitles.srt")
		if srt == nil {
			file, err := os.Create(filepath + m.Title + " Subtitles.srt")

			if err != nil {
				log.Panic("Unable to create subtitle file", err)
			}

			defer file.Close()
			defer en.Close()

			_, err = io.Copy(file, &en)
			if err != nil {
				log.Panic(err)
			}
			fmt.Println("Subtitles downloaded for:", dir)
			continue
		}

		fmt.Println("Subtitles already exist for:", dir)
	}
	fmt.Print("Finished - press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
