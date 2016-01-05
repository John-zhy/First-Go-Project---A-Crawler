package crawl

import (
	"golang.org/x/net/html"
	"strings"
	"fmt"
	"sync"
	"time"
	"encoding/csv"
	"os"
)
type AnimeConventionItem struct {
	ConventionItem
	atDoorRateALL string
	atDoorRateThu string
	atDoorRateFri string
	atDoorRateSat string
	atDoorRateSun string
}

func (item *AnimeConventionItem) readDescription(t *html.Tokenizer) {
	t.Next()
	item.description = string(t.Text())
}

func (item *AnimeConventionItem) readLatitude(t *html.Tokenizer) {
	_, val, _  := t.TagAttr()
	item.latitude = string(val)
}

func (item *AnimeConventionItem) readLongitude(t *html.Tokenizer) {
	_, val, _  := t.TagAttr()
	item.longitude = string(val)
}

func (item *AnimeConventionItem) readStartDate(t *html.Tokenizer) {
	_, val, _  := t.TagAttr()
	item.startDate = string(val)
}

func (item *AnimeConventionItem) readEndDate(t *html.Tokenizer) {
	_, val, _  := t.TagAttr()
	item.endDate = string(val)
}
func (item *AnimeConventionItem) readState(t *html.Tokenizer) {
	t.Next()
	item.state = string(t.Text())
}
func (item *AnimeConventionItem) readCity(t *html.Tokenizer) {
	t.Next()
	item.city = string(t.Text())
}

func (item *AnimeConventionItem) readCountry(t *html.Tokenizer, hasmore bool) {		
	if hasmore{
		_, val, _ := t.TagAttr()
		item.country = string(val)
		return
		
	}
	t.Next()
	item.country = string(t.Text())
}
func (item *AnimeConventionItem) readResgiterNowurl(t *html.Tokenizer) {	
	t.Next()
	if _, hasmore := t.TagName(); hasmore{
		if key, val, _  := t.TagAttr(); strings.EqualFold(string(key), "href") {
				item.registerNowURL = string(val)
		}
	}
}

func (item *AnimeConventionItem) readadvanceRate(t *html.Tokenizer) {	
	item.advanceRate = item.readRates(t)
}

func (item *AnimeConventionItem) readRates(t *html.Tokenizer) string {
	rates := ""
	for {
	 	label := t.Next()
	 	if label == html.EndTagToken{
	 		val, _:= t.TagName()
	 		if strings.EqualFold(string(val), "p"){
	 			break;
	 		}
	 	} 
		if label == html.TextToken{
			rates = strings.Join([]string{rates, string(t.Text())}, "\n")
		}
	}
	return strings.TrimSpace(rates)
}

func (item *AnimeConventionItem) readatDoorRate(t *html.Tokenizer) {	
	rates := item.readRates(t)
	item.atDoorRate = rates
	rateArray := strings.Split(rates, "\n")
	for  _, v := range(rateArray){
		//fmt.Println(v)
		if !strings.Contains(v, ":") {
			item.atDoorRateALL = v
			continue
		}
		tmp := strings.Split(v, ":")
		if strings.Contains(v, "All") || strings.Contains(v, "Both") {
			item.atDoorRateALL = tmp[1]
		}else if strings.Contains(v, "Thu"){
			item.atDoorRateThu = tmp[1]
		}else if strings.Contains(v, "Fri"){
			item.atDoorRateFri = tmp[1]
		}else if strings.Contains(v, "Sat"){
			item.atDoorRateSat = strings.Split(v, ":")[1]
		}else if strings.Contains(v, "Sun"){
			item.atDoorRateSun = strings.Split(v, ":")[1]
		}
	}
}

func (item *AnimeConventionItem) readNameAndLink(t *html.Tokenizer) {
	if label := t.Next();  label == html.StartTagToken{
		_, hasmore := t.TagName()
		if hasmore {
			if key, val, _  := t.TagAttr(); strings.EqualFold(string(key), "href") {
				item.siteURL = string(val)
			}
		}
	}
	if label := t.Next();  label == html.TextToken {
		item.name = string(t.Text())
	}
}

func (item *AnimeConventionItem) readLocation(t *html.Tokenizer) {
	for {
		if label := t.Next();  label == html.StartTagToken{
			_, hasmore := t.TagName()
			if hasmore {
				if _, val, _  := t.TagAttr(); strings.EqualFold(string(val), "name") {
					break
				}
			}
		}
	}
	if label := t.Next();  label == html.TextToken {
		item.location = string(t.Text())
	}
}

var mutex = &sync.Mutex{}

func (item *AnimeConventionItem) writeToCSV(writer *csv.Writer){
	mutex.Lock()
 	records := [][]string {
 		{item.name, item.location, item.city, 
 		item.state, item.country, item.startDate, 
 		item.endDate, item.latitude, item.longitude,
 		item.advanceRate, item.atDoorRate, item.atDoorRateALL, item.atDoorRateThu,
 		item.atDoorRateFri, item.atDoorRateSat, item.atDoorRateSun,
 		item.siteURL, item.registerNowURL, item.description},
 	}
	for _, record := range records {
        err := writer.Write(record)
        if err != nil {
          	fmt.Errorf("Error:\n", err)
            return
        }
    }
    mutex.Unlock()
}



func (item *AnimeConventionItem) Parse(t *html.Tokenizer) {
	for {
		label := t.Next()
		switch label {
			case html.ErrorToken:
				fmt.Errorf("%v\n", t.Err())
				return
			case html.TextToken: 
				switch string(t.Text()){
					case "Advance Rates:":
						//fmt.Println("rate")
						item.readadvanceRate(t)
					case "At-Door Rates:":
						item.readatDoorRate(t)
				}
			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				tag, hasmore := t.TagName()
				if strings.EqualFold(string(tag), "big"){
					item.readResgiterNowurl(t)
				}else if hasmore {
					key, val, hasmore  := t.TagAttr()
					if strings.EqualFold(string(key), "itemprop"){
						//fmt.Println(string(val))				
						switch string(val){
							case"description":
								item.readDescription(t)
							case "latitude":
								item.readLatitude(t)
							case "longitude":
								item.readLongitude(t)
							case "startDate":
								item.readStartDate(t)
							case "endDate":
								item.readEndDate(t)
							case "location":
								item.readLocation(t)
							case "addressLocality":
								item.readCity(t)
							case "addressRegion":
								item.readState(t)
							case "addressCountry":
								item.readCountry(t, hasmore)
							case "name":
								item.readNameAndLink(t)
						}
					}
				}
		}
	}
}
func (item *AnimeConventionItem) crawlInformation(list *ConventionList){
	for i := 0; i < 20 && strings.EqualFold(item.name , ""); i++ {
		time.Sleep(500 * time.Millisecond)
		crawlInformation(item.url, item)
	}
	//item.println()
	item.writeToCSV(csvwriter)
	item.finish(list)
}
type AnimeConventionList struct {
	ConventionList
}

var csvwriter *csv.Writer

func (list *AnimeConventionList) CrawlInformation(isFinish chan bool){
   	csvfile, err := os.Create("AnimeConvention.csv")
   	if err != nil {
   		fmt.Println("Error:", err)
   		return
    }
    defer csvfile.Close()
    csvwriter = csv.NewWriter(csvfile)
	// Headers
    var new_headers = []string {
     "name", "location", "city", 
     "state", "country", "start date", 
     "end date", "latitude", "longtitude", 
     "advance rates", "atDoorRate", "All", "Thu",
     "Fri", "Sat", "Sun",
     "site url", "register now url", "description"}
    returnError := csvwriter.Write(new_headers)
    if returnError != nil {
        fmt.Println(returnError)
    }
    list.url = "http://animecons.com/events/state.shtml/002800799"
	crawlInformation(list.url, list)
	list.waitFinished()
	csvwriter.Flush()
	isFinish <- true
}

func (list *AnimeConventionList) Parse(t *html.Tokenizer) {
	for {
		next := t.Next()
		switch next {
			case html.ErrorToken:
				return
			case html.TextToken:
				//fmt.Println(string(t.Text()))
			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				_, hasmore := t.TagName()
				if hasmore {
					key, val, _  := t.TagAttr()
					if strings.EqualFold(string(key), "href") && strings.HasPrefix(string(val), "/events/info.shtml"){
						var item = &AnimeConventionItem{}
						item.url = strings.Join([]string{"http://animecons.com", string(val)}, "")
						list.taskNum++
						go item.crawlInformation(&list.ConventionList)
						//time.Sleep(100 * time.Millisecond)
					}
				}
		}
	}
}