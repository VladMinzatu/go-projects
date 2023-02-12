package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
)

func TestHNServiceInvalidInput(t *testing.T) {

	var service HNService = NewHNService(NewTopStoriesRepoStub(nil, nil))

	t.Run("number of requested stories is too small", func(t *testing.T) {
		n := minStories - 1
		_, err := service.GetTopStories(n)
		want := ErrInvalidInput

		if err != want {
			t.Errorf("got %q want %q", err, want)
		}
	})

	t.Run("number of requested stories is too large", func(t *testing.T) {
		n := maxStories + 1
		_, err := service.GetTopStories(n)
		want := ErrInvalidInput

		if err != want {
			t.Errorf("got %q want %q", err, want)
		}
	})
}

func TestHNServiceHandlingRepoResult(t *testing.T) {

	t.Run("propagate the error from the repository", func(t *testing.T) {
		stubError := errors.New("There was an error fetching stories from the repo")
		service := NewHNService(NewTopStoriesRepoStub(nil, stubError))
		n := maxStories - 1

		_, err := service.GetTopStories(n)

		if err != stubError {
			t.Errorf("got %q want %q", err, stubError)
		}
	})

	t.Run("propagate the stories from the repository when there is no error and the result is nil", func(t *testing.T) {
		service := NewHNService(NewTopStoriesRepoStub(nil, nil))
		n := maxStories - 1

		stories, _ := service.GetTopStories(n)

		if stories != nil {
			t.Errorf("Expected to get a nil result")
		}
	})

	t.Run("propagate the stories from the repository when there is no error and the result is not nil", func(t *testing.T) {
		stubStories := []domain.Story{{Title: "New Go Features", Url: "https://www.go.com"}, {Title: "New ML model", Url: "https://www.ml.com"}}
		service := NewHNService(NewTopStoriesRepoStub(stubStories, nil))
		n := maxStories - 1

		stories, _ := service.GetTopStories(n)

		if !reflect.DeepEqual(stories, stubStories) {
			t.Errorf("expected to get the exact input stories propagated but got %q instead of %q", stories, stubStories)
		}
	})
}

type TopStoriesRepoStub struct {
	stories []domain.Story
	err     error
}

func (stub TopStoriesRepoStub) GetTopStories(n int) ([]domain.Story, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	return stub.stories, nil
}

func NewTopStoriesRepoStub(stories []domain.Story, err error) TopStoriesRepoStub {
	return TopStoriesRepoStub{stories: stories, err: err}
}
