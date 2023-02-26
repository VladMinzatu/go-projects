package adapters

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
)

type StoryResponseDto struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type HackerNewsClientImpl struct {
	client             *http.Client
	topStoriesUrl      string
	storyResolutionUrl string
}

func NewHackerNewsClient(client *http.Client, topStoriesUrl, storyResolutionUrl string) *HackerNewsClientImpl {
	return &HackerNewsClientImpl{client: client, topStoriesUrl: topStoriesUrl, storyResolutionUrl: storyResolutionUrl}
}

func (hnClient *HackerNewsClientImpl) GetTopStoryIds() ([]int, error) {
	var storyIds []int
	err := hnClient.fetchDataAndUnmarshal(hnClient.topStoriesUrl, &storyIds)
	return storyIds, err
}

func (hnClient *HackerNewsClientImpl) ResolveStory(id int) (domain.Story, error) {
	url := hnClient.storyResolutionUrl + strconv.Itoa(id) + ".json"
	var story StoryResponseDto
	err := hnClient.fetchDataAndUnmarshal(url, &story)
	return domain.Story{Title: story.Title, Url: story.Url}, err
}

func (hnClient *HackerNewsClientImpl) fetchDataAndUnmarshal(url string, response any) error {
	r, err := hnClient.client.Get(url)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return errors.New(string("status not OK"))
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, response)
}
