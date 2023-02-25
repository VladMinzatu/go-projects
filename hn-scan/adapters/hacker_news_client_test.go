package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
)

func TestFetchingTopStories(t *testing.T) {

	t.Run("unmarshals correct story id data successfully", func(t *testing.T) {
		testData := []string{"1", "2", "3", "4", "5"}
		ts, err := startTestServer(testData)
		if err != nil {
			t.Errorf("Failed to start test server due to failure: %s", err.Error())
		}
		defer ts.Close()

		client := NewClient(ts.URL)
		storyIds, err := client.GetTopStoryIds()
		if err != nil {
			t.Errorf("Error making request: %s", err.Error())
		}
		if !reflect.DeepEqual(storyIds, testData) {
			t.Errorf("Got unexpected data in response: %q", storyIds)
		}
	})

	t.Run("handles error codes by propagating errors", func(t *testing.T) {
		testHandlingNonOKStatusCodesByPropagatingErrors(t, callGetTopStoryIds)
	})

	t.Run("handles timeouts by propagating errors", func(t *testing.T) {
		testHandlingTimeoutsByPropagatingErrors(t, callGetTopStoryIds)
	})
}

func TestResolvingStoriesById(t *testing.T) {
	t.Run("unmarshals story data successfully", func(t *testing.T) {
		testData := StoryResponseDto{"Go article", "www.go.com"}
		ts, err := startTestServer(testData)
		if err != nil {
			t.Errorf("Failed to start test server due to failure: %s", err.Error())
		}

		client := NewClient(storyResolutionUrlPrefix(ts.URL))
		story, err := client.ResolveStory("123")
		if err != nil {
			t.Errorf("Error making request: %s", err.Error())
		}
		if !reflect.DeepEqual(story, domain.Story{Title: testData.Title, Url: testData.Url}) {
			t.Errorf("Got unexpected data in response: %q", story)
		}
	})

	t.Run("passing the correct url to resolve a story", func(t *testing.T) {
		id := "the_correct_test_id"
		ts := startTestServerExpectingTheCorrectStoryUrl(id)
		defer ts.Close()

		client := NewClient(storyResolutionUrlPrefix(ts.URL))
		story, err := client.ResolveStory(id)
		if err != nil {
			t.Errorf("Error making request: %s", err.Error())
		}
		if story.Title != "Well done" {
			t.Errorf("Got unexpected data in response: %q", story)
		}
	})

	t.Run("handles error codes by propagating errors", func(t *testing.T) {
		testHandlingNonOKStatusCodesByPropagatingErrors(t, callResolveRandomStory)
	})

	t.Run("handles timeouts by propagating errors", func(t *testing.T) {
		testHandlingTimeoutsByPropagatingErrors(t, callResolveRandomStory)
	})
}

func storyResolutionUrlPrefix(url string) string {
	return url + "/?story_id="
}

func callGetTopStoryIds(client *HackerNewsClientImpl) (any, error) {
	return client.GetTopStoryIds()
}

func callResolveRandomStory(client *HackerNewsClientImpl) (any, error) {
	return client.ResolveStory("123")
}

type callExportedClientMethod func(*HackerNewsClientImpl) (any, error)

func testHandlingNonOKStatusCodesByPropagatingErrors(t *testing.T, fn callExportedClientMethod) {
	ts := startTestServerWith404Response()
	defer ts.Close()
	client := NewClient(storyResolutionUrlPrefix(ts.URL))

	_, err := fn(client)
	if err == nil {
		t.Errorf("Expected error when server responds with 404, but didn't get one. Payload")
	}
	if err.Error() != "status not OK" {
		t.Errorf("Wrong error message. Non-200 return status code not handled as expected")
	}
}

func testHandlingTimeoutsByPropagatingErrors(t *testing.T, fn callExportedClientMethod) {
	ts := startTestServerOverloaded()
	defer ts.Close()

	client := NewClient(ts.URL)
	start := time.Now()
	_, err := fn(client)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected an error but got nil")
	}
	if elapsed > time.Second+100*time.Millisecond {
		t.Errorf("Request took too long: %v", elapsed)
	}
}

func NewClient(url string) *HackerNewsClientImpl {
	client := http.Client{Timeout: 1 * time.Second}
	return NewHackerNewsClient(&client, url, url)
}

func startTestServer(data any) (*httptest.Server, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonData)
			}))
	return ts, nil
}

func startTestServerWith404Response() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Error(
					w, "The resource was not found", http.StatusNotFound)
			}))
}

func startTestServerOverloaded() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				http.Error(w, "I have failed you, but you should have given up by now :(", http.StatusInternalServerError)
			}))
}

func startTestServerExpectingTheCorrectStoryUrl(id string) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.String(), id) {
					http.Error(w, "The resource was not found", http.StatusNotFound)
				}
				story, err := json.Marshal(StoryResponseDto{Title: "Well done", Url: "www"})
				if err != nil {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(story)
			}))
}
