package crawl

import (
	//"sync"
	"fmt"
	"net/http"
	"time"
	"golang.org/x/net/html"
	"sync/atomic"
	"github.com/kellydunn/golang-geo"
	"strconv"
	"strings"
	"encoding/csv"
)

type ConventionItem struct{
	url string
	description string //eventDescriptions
	latitude string //eventLatitude
	longitude string //eventLongitude
	startDate string //eventStartDate
	endDate string //eventEndDate
	name string //eventName
	location string //eventVenue
	city string //eventCity
	state string //eventState
	country string
	registerNowURL string
	siteURL string
	address string //eventAddress
	zip string //eventZip
	advanceRate	string 
	atDoorRate	string
	status	string
	artistAlleyBoothPrice string
	exhibitorBoothPrice	string 
	artistsAlleyRegistrationUrl	string
	exhibitorsBoothRegistrationUrl	string 
	artistAlleySpotsAvailablility string
	boothOpenEnrollmentDay	string 
	boothEnrollmentEndDay	string
	organizerContactUrl string
	organizerContactInfo string
	eventType string								
}

type TokenParser interface {
	Parse(t *html.Tokenizer)
}

func (list *ConventionList) waitFinished() {
	for list.taskNum > atomic.LoadInt32(&list.finishNum) {
		//fmt.Printf("finish = %d task = %d\n", atomic.LoadInt32(&list.finishNum), list.taskNum)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("finish = %d task = %d\n", atomic.LoadInt32(&list.finishNum), list.taskNum)
	
}

type ConventionList struct {
	url string
	taskNum int32
	finishNum int32
} 

func crawlInformation(url string, parser TokenParser) {
	//fmt.Println("crawling", url)
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
    	Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Errorf("fail to get\n")
		return
	}
	t := html.NewTokenizer(resp.Body)
	defer resp.Body.Close()
	parser.Parse(t)
}


func (item *ConventionItem) println() {
	mutex.Lock()
	fmt.Println(item.url)
	fmt.Println("name\tlocation\tstart date\tend date\tlatitude\tlontitude")
	fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", item.name, item.location, item.startDate, item.endDate, item.latitude, item.longitude)
	fmt.Println("description:")
	fmt.Println(item.description)
	mutex.Unlock()
}

func (item *ConventionItem) finish(list *ConventionList) {
	atomic.AddInt32(&list.finishNum, 1)
}

func (item *ConventionItem) setGeoPoint() bool{
	var address string
	loc := item.location
	if strings.Contains(item.location,",") {
		loc = strings.TrimSpace(strings.Split(item.location, ",")[0])
	}
	addrs := make([]string, 0)
	for _, addr := range([]string{loc, item.city, item.state, item.country}){
		if len(addr) >= 1{
			addrs = append(addrs, addr)
		}
	}
	address = strings.Join(addrs[:], ", ")
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
    	Timeout: timeout,
	}
	query := &geo.GoogleGeocoder{&client}
	point, err:= query.Geocode(address)
	if err != nil {
		fmt.Errorf("cannot find geopoint addr = %s err = %v \n", address, err)
		return false
	}
	item.latitude = strconv.FormatFloat(point.Lat(), 'f', 6, 64)
	item.longitude = strconv.FormatFloat(point.Lng(), 'f', 6, 64)
	return true
	//fmt.Println(address, item.latitude, item.longitude)
}


func (item *ConventionItem) writeToCSV(writer *csv.Writer){
	mutexComicConvention.Lock()
		
	
 	records := [][]string {
 		{item.name, item.location, item.address, item.zip,
 		item.city, item.state, item.country, item.latitude, item.longitude,
 		item.startDate, item.endDate, item.advanceRate, item.atDoorRate,
 		item.siteURL, item.registerNowURL, item.description,
 		item.status, item.artistAlleyBoothPrice, item.exhibitorBoothPrice,
 		item.artistsAlleyRegistrationUrl, item.exhibitorsBoothRegistrationUrl,
 		item.artistAlleySpotsAvailablility, item.boothOpenEnrollmentDay,
 		item.boothEnrollmentEndDay,	item.organizerContactUrl,
 		item.organizerContactInfo, item.eventType},
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
