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

var (
	// regex for a "title of movie (year)"
	moviere, _ = regexp.Compile(`^(.+)\((\d{4})\)`)
	path       = `./`
)

func main() {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	// loop our dir
	for _, file := range files {
		if file.IsDir() && moviere.MatchString(file.Name()) {
			innerfiles, err := ioutil.ReadDir(path + file.Name())
			if err != nil {
				return
			}
			// if subtitles exist already, skip
			hassub := false
			for _, infile := range innerfiles {
				if strings.HasSuffix(infile.Name(), ".srt") {
					fmt.Println("Found subtitle in: ", file.Name(), ", skipping")
					hassub = true
					break
				}
			}
			if hassub {
				continue
			}
			movie := moviere.FindStringSubmatch(file.Name())
			title, year := movie[1], movie[2]
			title = strings.TrimSpace(title)
			title = strings.Replace(title, " ", "+", -1)
			query := &gomdb.QueryData{Title: title, Year: year}
			res, err := gomdb.Search(query)
			if err != nil {
				fmt.Println("Movie not found: ", movie[1])
				continue
			}
			subs, err := yifysubs.GetSubtitles(res.Search[0].ImdbID)
			if err != nil {
				log.Panic(err)
			}
			index, max := 0, subs["english"][0].Rating
			for i := range subs["english"] {
				if subs["english"][i].Rating > max {
					max = subs["english"][i].Rating
					index = i
				}
			}
			// get the extact title of the movie
			moviefilename := ""
			for _, infile := range innerfiles {
				if strings.HasSuffix(infile.Name(), ".mp4") || strings.HasSuffix(infile.Name(), ".mkv") {
					moviefilename = infile.Name()
					moviefilename = moviefilename[:len(moviefilename)-4]
				}
			}
			// download, copy the subtitles and
			finalsub := subs["english"][index]
			file, err := os.Create(path + file.Name() + `\` + moviefilename + ".srt")
			if err != nil {
				log.Panic(err)
			}

			defer file.Close()
			defer finalsub.Close()

			_, err = io.Copy(file, &finalsub)
			if err != nil {
				log.Panic(err)
			}

		}
	}

}
