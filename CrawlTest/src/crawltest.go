package main

import (
	"fmt"
	"crawl"
)
func crawlAnimeConvention() {
	var list = &crawl.AnimeConventionList{}
	c := make(chan bool)
	go list.CrawlInformation(c)	
	isfinish := <- c
	fmt.Println(isfinish)
}

func crawlComicConvention() {
	var list = &crawl.ComicConventionList{}
	c := make(chan bool)
	go list.CrawlInformation(c)	
	isfinish := <- c
	fmt.Println(isfinish)
}

func crawlConventionScene() {
	var list = &crawl.ConventionSceneList{}
	c := make(chan bool)
	go list.CrawlInformation(c)	
	isfinish := <- c
	fmt.Println(isfinish)
}

func main() {
	crawlConventionScene()
}

