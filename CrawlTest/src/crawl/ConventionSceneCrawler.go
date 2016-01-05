package crawl

import (
	"fmt"
	"strings"
	"golang.org/x/net/html"
	"os"
	"encoding/csv"
	"sync"
	"time"
)

type ConventionSceneItem struct {
	ConventionItem
}
type ConventionSceneList struct {
	ConventionList
}

var mutexConventionScene = &sync.Mutex{}

var conventionScenenWriter *csv.Writer
func (item *ConventionSceneItem) writeToCSV(writer *csv.Writer){
 	records := [][]string {
 		{item.name, item.location, item.city, 
 		item.state, item.country, item.startDate, 
 		item.endDate, item.latitude, item.longitude,
 		item.siteURL, item.registerNowURL, item.description},
 	}
 	mutexConventionScene.Lock()
	for _, record := range records {
        err := writer.Write(record)
        if err != nil {
          	fmt.Errorf("Error:\n", err)
            return
        }
    }
    mutexConventionScene.Unlock()
}


func (list *ConventionSceneList) CrawlInformation(isFinish chan bool){
   	csvfile, err := os.Create("ConventionScene.csv")
   	if err != nil {
   		fmt.Println("Error:", err)
   		return
    }
    defer csvfile.Close()
    conventionScenenWriter = csv.NewWriter(csvfile)
	// Headers
    var new_headers = []string {
     "name", "location", "city", 
     "state", "country", "start date", 
     "end date", "latitude", "longitude", 
     "site url", "register now url", "description"}
    returnError := conventionScenenWriter.Write(new_headers)
    if returnError != nil {
        fmt.Println(returnError)
    }
    list.url = "http://www.conventionscene.com/schedules/comicbookconventions/"
	crawlInformation(list.url, list)
	list.waitFinished()
	conventionScenenWriter.Flush()
	isFinish <- true
}

func (list *ConventionSceneList) Parse(t *html.Tokenizer) {
	item := &ConventionSceneItem{}
	for {
		label := t.Next()
		switch label {
			case html.ErrorToken:
				fmt.Errorf("%v\n", t.Err())
				return
			case html.TextToken: 

			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				_, hasmore := t.TagName()
				if hasmore {
					_, val, _  := t.TagAttr()
					if strings.EqualFold(string(val), "gigpress-row active gigpress-tour") ||
						 strings.EqualFold(string(val), "gigpress-row active gigpress-alt gigpress-tour"){
						list.taskNum++
						item.readConventionSceneItem(t)
					}else if strings.EqualFold(string(val), "gigpress-info active gigpress-tour")||
						strings.EqualFold(string(val), "gigpress-info active gigpress-alt gigpress-tour"){
						item.readLocation(t)		
						go item.complete(list)
						//fmt.Println(item.location, item.city, item.state)
						item = &ConventionSceneItem{}
					}
				}
		}
	}
}

func (item *ConventionSceneItem) complete(list *ConventionSceneList) {
	for i := 0; i < 25 && !item.setGeoPoint(); i++{
		time.Sleep(500 * time.Millisecond)
	}
	item.writeToCSV(conventionScenenWriter)
	item.finish(&list.ConventionList)
}

func (item *ConventionSceneItem) readConventionSceneItem(t *html.Tokenizer) {
	for {
		label := t.Next()
		switch label {
			case html.ErrorToken:
				fmt.Errorf("%v\n", t.Err())
				return
			case html.TextToken: 

			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				_, hasmore := t.TagName()
				if hasmore {
					key, val, _  := t.TagAttr()
					if strings.EqualFold(string(key), "class"){
						switch string(val) {
							case "gigpress-date":
								item.readDate(t)
							case "gigpress-city":
								item.readCityAndState(t)
							case "gigpress-venue":
									item.readNameAndLink(t)
							case "gigpress-country":
								item.readContry(t)
								return
						}
					}
				}
		}
	}
}

func (item *ConventionSceneItem) readDate(t *html.Tokenizer) {
	if label := t.Next(); label == html.TextToken {
		val := string(t.Text())
		if strings.Contains(val, "-"){
			dates := strings.Split(val, "-")
			item.startDate = item.formatDate(dates[0])
			item.endDate = item.formatDate(dates[1])
		} else {
			item.startDate = item.formatDate(val)
		}
	}
	//fmt.Println(item.startDate)
}

func (item *ConventionSceneItem) formatDate(date string) string{
	dates := strings.Split(strings.TrimSpace(date), "/")
	return dates[0] + "/" + dates[1] + "/20" + dates[2]
}

func (item *ConventionSceneItem) readCityAndState(t *html.Tokenizer) {
	if label := t.Next(); label == html.TextToken {
		val := string(t.Text())
		if strings.Contains(val, ","){
			addr := strings.Split(val, ",")
			item.city = strings.TrimSpace(addr[0])
			item.state = strings.TrimSpace(addr[1])
		} else {
			item.city = strings.TrimSpace(val)
		}
	}
	//fmt.Println(item.city)
}

func (item *ConventionSceneItem) readContry(t *html.Tokenizer) {
	if label := t.Next(); label == html.TextToken {
		item.country = string(t.Text())
	}
	//fmt.Println(item.country)
}
func (item *ConventionSceneItem) readNameAndLink(t *html.Tokenizer) {
	if label := t.Next(); label == html.StartTagToken {
		if _, hasmore := t.TagName(); hasmore{
			 if key, val, _  := t.TagAttr(); strings.EqualFold(string(key), "href") {
			 	item.siteURL = string(val)
			 }
		}
		if label := t.Next(); label == html.TextToken {
			item.name = string(t.Text())
		}
	}
	//fmt.Println(item.siteURL)
}

func (item *ConventionSceneItem) readLocation(t *html.Tokenizer) {
	for {
		label := t.Next()
		switch label {
			case html.ErrorToken:
				fmt.Errorf("%v\n", t.Err())
				return
			case html.StartTagToken:
				if _, hasmore := t.TagName(); hasmore {		
					if _, _, hashmore  := t.TagAttr(); hashmore{
							if _, val, _ := t.TagAttr(); strings.EqualFold(string(val), "gigpress-address"){
								if label := t.Next(); label == html.TextToken{
									item.location = strings.TrimSpace(string(t.Text()))
									return
								}
							}
					}
				}
			case html.EndTagToken:
				if tag, _ := t.TagName(); strings.EqualFold(string(tag), "tr"){
					return
				}
		}
	}
}