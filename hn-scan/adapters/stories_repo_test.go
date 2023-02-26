package adapters

import (
	"errors"
	"reflect"
	"testing"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
)

func TestGettingTopStoriesHappyPath(t *testing.T) {
	storyResolutions := map[string]domain.Story{
		"1": {Title: "New Go features", Url: "www.go.com"},
		"2": {Title: "New Tensorflow realease", Url: "www.tensorflow.com"},
		"3": {Title: "What's new in Java", Url: "www.java.com"},
		"4": {Title: "What's new in Python", Url: "www.python.com"},
		"5": {Title: "Breaking world news", Url: "www.news.com"},
	}
	mockClient := &MockHackerNewsClient{
		storyIds:         []string{"1", "2", "3", "4", "5"},
		storyIdsError:    nil,
		storyResolutions: storyResolutions,
	}
	expected := []domain.Story{storyResolutions["1"], storyResolutions["2"], storyResolutions["3"]}
	repo := NewTopStoriesRepo(mockClient)
	response, _ := repo.GetTopStories(3)

	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Got unexpected data in response: %q", response)
	}
}

func TestWillNotTryToResolveMoreStoriesThanTheLimit(t *testing.T) {
	storyResolutions := map[string]domain.Story{
		"1": {Title: "New Go features", Url: "www.go.com"},
		"2": {Title: "New Tensorflow realease", Url: "www.tensorflow.com"},
		"3": {Title: "What's new in Java", Url: "www.java.com"},
		// only 3 resolvable
	}
	mockClient := &MockHackerNewsClient{
		storyIds:         []string{"1", "2", "3", "4", "5"}, // but 5 story ids returned
		storyIdsError:    nil,
		storyResolutions: storyResolutions,
	}
	expected := []domain.Story{storyResolutions["1"], storyResolutions["2"], storyResolutions["3"]}
	repo := NewTopStoriesRepo(mockClient)
	response, _ := repo.GetTopStories(3) // it's ok because we only try to resolve the first 3

	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Got unexpected data in response: %q", response)
	}
}

func TestStoriesThatCannotBeResolvedAreSkipped(t *testing.T) {
	storyResolutions := map[string]domain.Story{
		"1": {Title: "New Go features", Url: "www.go.com"},
		"3": {Title: "What's new in Java", Url: "www.java.com"},
	}
	mockClient := &MockHackerNewsClient{
		storyIds:         []string{"1", "2", "3", "4", "5"},
		storyIdsError:    nil,
		storyResolutions: storyResolutions,
	}
	expected := []domain.Story{storyResolutions["1"], storyResolutions["3"]} // 2 is simply skipped in the result
	repo := NewTopStoriesRepo(mockClient)
	response, _ := repo.GetTopStories(3)

	if !reflect.DeepEqual(response, expected) {
		t.Errorf("Got unexpected data in response: %q", response)
	}
}

func TestFailureToGetStoryIds(t *testing.T) {
	storyResolutions := map[string]domain.Story{
		"1": {Title: "New Go features", Url: "www.go.com"},
		"2": {Title: "New Tensorflow realease", Url: "www.tensorflow.com"},
		"3": {Title: "What's new in Java", Url: "www.java.com"},
	}
	errorFetchingTopStoryIds := errors.New("Failed to fetch top story ids")
	mockClient := &MockHackerNewsClient{
		storyIds:         nil,
		storyIdsError:    errorFetchingTopStoryIds,
		storyResolutions: storyResolutions,
	}
	expected := errorFetchingTopStoryIds
	repo := NewTopStoriesRepo(mockClient)
	result, err := repo.GetTopStories(3)

	if expected != err {
		t.Errorf("Call expected to return specific error, but didn't. Instead go result=%q and err=%q", result, err)
	}
}

type MockHackerNewsClient struct {
	storyIds         []string
	storyIdsError    error
	storyResolutions map[string]domain.Story
}

func (client *MockHackerNewsClient) GetTopStoryIds() ([]string, error) {
	if client.storyIdsError != nil {
		return nil, client.storyIdsError
	}
	return client.storyIds, nil
}

func (client *MockHackerNewsClient) ResolveStory(id string) (domain.Story, error) {
	if result, found := client.storyResolutions[id]; found {
		return result, nil
	}
	return domain.Story{}, errors.New("Story could not be resolved by mock")
}
