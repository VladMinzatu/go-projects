package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/VladMinzatu/go-projects/hn-scan/adapters"
	"github.com/VladMinzatu/go-projects/hn-scan/core/service"
)

func main() {
	svc := setUpService()
	request, _ := service.NewHNServiceRequest(50, []string{"go"})
	result, err := svc.GetTopStories(request)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	fmt.Println(result)
}

func setUpService() service.HNService {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	topStoriesUrl := "https://hacker-news.firebaseio.com/v0/topstories.json"
	storyResolutionUrl := "https://hacker-news.firebaseio.com/v0/item/"
	hnClient := adapters.NewHackerNewsClient(&httpClient, topStoriesUrl, storyResolutionUrl)
	return service.NewHNService(adapters.NewTopStoriesRepo(hnClient))
}
