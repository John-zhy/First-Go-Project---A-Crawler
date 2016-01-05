package crawl

import (
	"fmt"
	"strings"
	"golang.org/x/net/html"
	"time"
	"os"
	"sync"
	"encoding/csv"
)
type ComicConventionItem struct {
	ConventionItem
	genres string
}
type ComicConventionList struct {
	ConventionList
}
var mutexComicConvention = &sync.Mutex{}

func (item *ComicConventionItem) writeToCSV(writer *csv.Writer){
	mutexComicConvention.Lock()
 	records := [][]string {
 		{item.name, item.location, item.city, 
 		item.state, item.country, item.startDate, 
 		item.endDate, item.latitude, item.longitude,
 		item.genres,
 		item.siteURL, item.registerNowURL, item.description},
 	}
	for _, record := range records {
        err := writer.Write(record)
        if err != nil {
          	fmt.Errorf("Error:\n", err)
            return
        }
    }
    mutexComicConvention.Unlock()
}

var comicConventionWriter *csv.Writer

func (item *ComicConventionItem) crawlInformation(list *ConventionList){
	for i := 0; i < 20 && strings.EqualFold(item.name , ""); i++ {
		time.Sleep(500 * time.Millisecond)
		crawlInformation(item.url, item)
	}
	//item.println()
	item.setGeoPoint()
	item.writeToCSV(comicConventionWriter)
	item.finish(list)
}

func (list *ComicConventionList) CrawlInformation(isFinish chan bool){
   	csvfile, err := os.Create("ComicConvention.csv")
   	if err != nil {
   		fmt.Println("Error:", err)
   		return
    }
    defer csvfile.Close()
    comicConventionWriter = csv.NewWriter(csvfile)
	// Headers
    var new_headers = []string {
     "name", "location", "city", 
     "state", "country", "start date", 
     "end date", "latitude", "longitude", 
     "genres",
     "site url", "register now url", "description"}
    returnError := comicConventionWriter.Write(new_headers)
    if returnError != nil {
        fmt.Println(returnError)
    }
    list.url = "http://www.upcomingcons.com/comic-conventions"
	crawlInformation(list.url, list)
	list.waitFinished()
	isFinish <- true
}

func (list *ComicConventionList) Parse(t *html.Tokenizer) {
	for {
		label := t.Next()
		switch label {
			case html.ErrorToken:
				fmt.Errorf("%v\n", t.Err())
				return
			case html.TextToken: 

			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				_, hasmore := t.TagName()
				key, val, hasmore  := t.TagAttr()
				//fmt.Printf("%s \t %s\n", string(key), string(val))
				if strings.EqualFold(string(key), "itemprop") && strings.EqualFold(string(val), "url") && hasmore{
					key, val, _  = t.TagAttr()
					if strings.EqualFold(string(key), "href") {					
					 	var item = &ComicConventionItem{}
					 	item.url = strings.Join([]string{"http://www.upcomingcons.com", string(val)}, "")
					 	list.taskNum++
					 	//fmt.Println(item.url)
					 	go item.crawlInformation(&list.ConventionList)
					}
				}
		}
	}
}

func (item *ComicConventionItem) Parse(t *html.Tokenizer) {
	for {
		label := t.Next()
		switch label {
			case html.ErrorToken:
				fmt.Errorf("%v\n", t.Err())
				return
			case html.TextToken: 
				if strings.EqualFold(string(t.Text()), "Genres") {
					item.readGenres(t)
				}
			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				tag, hasmore := t.TagName()
				if hasmore{
					if strings.EqualFold(string(tag), "h1"){
						item.readName(t)
						continue
					}
					key, val, _  := t.TagAttr()
					if strings.EqualFold(string(val), "location") && strings.EqualFold(string(key), "id"){
						item.readCityStateCounty(t)
					}else if strings.EqualFold(string(key), "itemprop"){
						switch string(val) {
							case "startDate":
								item.readStartDate(t)
							case "endDate":
								item.readEndDate(t)
							case "name":
								item.readLocation(t)
							case "addressRegion":
								item.readSiteLink(t)
						}
					}
				}
		}
	}
}



func (item *ComicConventionItem) readName(t *html.Tokenizer) {
	if label := t.Next();  label == html.TextToken {
		item.name = strings.TrimSpace(string(t.Text()))
	}
}

func (item *ComicConventionItem) readStartDate(t *html.Tokenizer) {
	_, val, _  := t.TagAttr()
	item.startDate = string(val)
}

func (item *ComicConventionItem) readEndDate(t *html.Tokenizer) {
	_, val, _  := t.TagAttr()
	item.endDate = string(val)
}

func (item *ComicConventionItem) readLocation(t *html.Tokenizer) {
	if label := t.Next(); label == html.TextToken{
		item.location = string(t.Text())
	}
}

func (item *ComicConventionItem) readSiteLink(t *html.Tokenizer) {
	for {
		if label := t.Next(); label == html.StartTagToken{
			if _, hasmore := t.TagName(); hasmore{
				key, val, _ := t.TagAttr()
				if strings.EqualFold(string(key), "href") {
					item.siteURL = string(val)
					break
				}
			}
		}
	}
	//fmt.Println(item.siteURL)
}

func (item *ComicConventionItem) readGenres(t *html.Tokenizer) {
	t.Next()
	item.genres = ""
	for {
		label := t.Next(); 
		if label == html.EndTagToken {
			if tag, _ := t.TagName(); strings.EqualFold(string(tag), "div"){
				break
			}
		}else if label == html.TextToken{
			if genre := strings.TrimSpace(string(t.Text())); len(genre) > 1 {
				item.genres = strings.Join([]string{item.genres, genre}, "\n")
			}
		}
	}
	item.genres = strings.TrimSpace(item.genres)
	//fmt.Println(item.genres)
}

func (item *ComicConventionItem) readCityStateCounty(t *html.Tokenizer) {
	if label := t.Next(); label == html.TextToken{
		address := strings.Split(string(t.Text()), ",")
		switch {
			case len(address) > 2: 
				item.country = strings.TrimSpace(address[2])
				fallthrough
			case len(address) == 2:
				item.state = strings.TrimSpace(address[1])
				fallthrough
			case len(address) == 1:
				item.city = strings.TrimSpace(address[0])
		}
		if strings.EqualFold(item.country, ""){
			item.country = "USA"
		}
		//fmt.Printf("%s,%s,%s\n", item.country, item.state, item.city)
	}
}