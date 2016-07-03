package main

import (
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

// scan a folder for dirs and return them in a list
func scandirs(path string) []string {
	r, _ := regexp.Compile(`.+\)`)
	count := 0
	dirs := 0
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Panicln("Reading path failed: ", err)
	}
	// this is ugly, find a better way so this doesn't loop dirs twice
	for _, f := range files {
		if f.IsDir() && r.MatchString(f.Name()) && f.Name()[0] != '(' {
			dirs++
		}
	}
	founddirs := make([]string, dirs, dirs+1)
	for _, f := range files {
		if f.IsDir() && r.MatchString(f.Name()) && f.Name()[0] != '(' {
			founddirs[count] = f.Name()
			count++
		}
	}
	return founddirs
}

// Get a raw dir title and extract information from it, should work for all YTS downloads
func getmovie(in string) (string, string) {
	in = strings.TrimSuffix(in, " [YTS.AG]")

	date := in[len(in)-5:]
	date = strings.TrimSuffix(date, ")")

	title := in[:strings.Index(in, "(")-1]

	return title, date
}

func getimdb(title string, date string) string {
	query := &gomdb.QueryData{Title: title, Year: date}
	res, err := gomdb.MovieByTitle(query)
	if err != nil {
		//fmt.Println("Querying OMDB failed:", err)
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
	fmt.Println("Finding movies in:", path)
	dirs := scandirs(path)
	m := new(movie)

	for _, dir := range dirs {
		m.Title, m.Date = getmovie(dir)

		subs, err := yifysubs.GetSubtitles(getimdb(m.Title, m.Date))
		if err != nil {
			//fmt.Println("Sub finding failed")
			continue
		}
		en := subs["english"][0]

		file, err := os.Create(path + `\` + dir + `\` + m.Title + " " + "subtitles.srt")
		if err != nil {
			log.Panic("Unable to create subtitle file", err)
		}

		defer file.Close()
		defer en.Close()

		_, err = io.Copy(file, &en)
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("Successfully got subtitles for:", dir)
	}
}
