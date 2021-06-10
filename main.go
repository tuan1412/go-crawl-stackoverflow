package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Question struct {
	Title string `json:"title"`
	Author string `json:"author"`
	UpdatedAt string `json:"updatedAt"`
	Tags []string `json:"tags"`
}

func (q Question) ToString() []string {
	return []string {
		q.Title,
		q.Author,
		q.UpdatedAt,
		strings.Join(q.Tags, ","),
	}
}
func crawl (url string) ([]Question, error){
	questions := []Question{}

	res, err := http.Get(url)
	if err != nil {
		return questions, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
    return questions, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
  }

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return questions, err
	}


	doc.Find(".question-summary .summary").Each(func (i int , questionSec *goquery.Selection) {
		title:= questionSec.Find(".question-hyperlink").First().Text()
		author := questionSec.Find(".user-details a").First().Text()
		updatedAt, _ := questionSec.Find(".user-action-time span").First().Attr("title")
		tags := []string{}

		questionSec.Find(".post-tag").Each(func (i int, tagSec *goquery.Selection) {
			tag := tagSec.Text()
			tags = append(tags, tag)
		})

		question := Question{
			Title: title,
			Author: author,
			UpdatedAt: updatedAt,
			Tags: tags,
		}

		questions = append(questions, question)
	})

	return questions, nil
}

func main() {
	fileName := "questions.csv"
	file, err := os.Create(fileName)
	if err != nil {
			log.Fatalf("Could not create %s", fileName)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	maxPage := 100
	maxGoRoutine := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(maxPage)

	urlChannel := make(chan int, maxPage)
	for i:= 1; i <= maxPage; i++ {
		urlChannel <- i
	}

	start := time.Now()

	for i:= 0; i < maxGoRoutine; i++ {
		go func () {
			for {
				page := <- urlChannel
				url := fmt.Sprintf("https://stackoverflow.com/questions?tab=newest&page=%d", page)
				questions, _ := crawl(url)

				// FUTURE: tách write file thành một goroutine riêng
				for _, q := range questions {
					writer.Write(q.ToString())
				}
				wg.Done()
			}
		
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("Time excute", elapsed)
}