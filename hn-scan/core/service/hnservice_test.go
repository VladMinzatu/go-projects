package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
)

func TestHNServiceHandlingRepoResult(t *testing.T) {

	t.Run("propagate the error from the repository", func(t *testing.T) {
		stubError := errors.New("There was an error fetching stories from the repo")
		service := NewHNService(NewTopStoriesRepoStub(nil, stubError))
		req, _ := NewHNServiceRequest(maxStories-1, nil)

		_, err := service.GetTopStories(req)

		if err != stubError {
			t.Errorf("got %q want %q", err, stubError)
		}
	})

	t.Run("propagate the stories from the repository when there is no error and the result is nil", func(t *testing.T) {
		service := NewHNService(NewTopStoriesRepoStub(nil, nil))
		req, _ := NewHNServiceRequest(maxStories-1, nil)

		stories, _ := service.GetTopStories(req)

		if stories != nil {
			t.Errorf("Expected to get a nil result")
		}
	})

	t.Run("propagate the stories from the repository when there is no error and the result is not nil", func(t *testing.T) {
		goStory := domain.Story{Title: "New Go Features", Url: "https://www.go.com"}
		mlStory := domain.Story{Title: "New ML model", Url: "https://www.ml.com"}
		javaStory := domain.Story{Title: "Design patterns in Java", Url: "https://www.java.com"}
		stubStories := []domain.Story{goStory, mlStory, javaStory}
		limit := len(stubStories)

		service := NewHNService(NewTopStoriesRepoStub(stubStories, nil))
		req, _ := NewHNServiceRequest(maxStories-1, nil)

		tests := []struct {
			input    *HNServiceRequest
			expected []domain.Story
		}{
			{
				&HNServiceRequest{limit: limit, terms: nil},
				stubStories,
			},
			{
				&HNServiceRequest{limit: limit, terms: []string{}},
				stubStories,
			},
			{
				&HNServiceRequest{limit: limit, terms: []string{"foo"}},
				nil,
			},
			{
				&HNServiceRequest{limit: limit, terms: []string{"go"}},
				[]domain.Story{goStory},
			},
			{
				&HNServiceRequest{limit: limit, terms: []string{"java", "go"}},
				[]domain.Story{goStory, javaStory},
			},
			{
				&HNServiceRequest{limit: limit, terms: []string{"go", "features"}},
				[]domain.Story{goStory},
			},
		}

		for _, test := range tests {
			result, err := service.GetTopStories(test.input)
			if err != nil {
				t.Errorf("Unexpected error in test case for input %q: Got error %v", test.input, err)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("GetTopStories(%q) = %v, but expected %v", test.input, result, test.expected)
			}
		}

		stories, _ := service.GetTopStories(req)

		if !reflect.DeepEqual(stories, stubStories) {
			t.Errorf("expected to get the exact input stories propagated but got %q instead of %q", stories, stubStories)
		}
	})
}

func TestHNServiceRequest(t *testing.T) {

	t.Run("return a request struct with the provided fields if all is valid", func(t *testing.T) {
		n := minStories + 1
		terms := []string{"foo", "bar"}
		req, err := NewHNServiceRequest(n, terms)

		if err != nil {
			t.Error("expected non-nil error, but got nil")
		}

		if !reflect.DeepEqual(req.Terms(), terms) {
			t.Errorf("expected terms to be %q but got %q", terms, req.Terms())
		}
		if req.Limit() != n {
			t.Errorf("expected limit to be %q but got %q", n, req.Limit())
		}
	})

	t.Run("return an error if the number of requested stories is too large", func(t *testing.T) {
		n := maxStories + 1
		_, err := NewHNServiceRequest(n, []string{})

		if err == nil {
			t.Error("expected non-nil error, but got nil")
		}
	})

	t.Run("return an error if the number of requested stories is too small", func(t *testing.T) {
		n := minStories - 1
		_, err := NewHNServiceRequest(n, []string{})

		if err == nil {
			t.Error("expected non-nil error, but got nil")
		}
	})

	t.Run("return a request struct with an empty list of terms if terms param is nil", func(t *testing.T) {
		n := minStories + 1
		req, err := NewHNServiceRequest(n, nil)

		if err != nil {
			t.Error("expected non-nil error, but got nil")
		}
		if req.Terms() == nil || len(req.Terms()) != 0 {
			t.Errorf("expected empty slice of terms, got %q", req.Terms())
		}
	})
}

func TestWordExtraction(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			"",
			nil,
		},
		{
			"   , !   ",
			nil,
		},
		{
			"Hello, world!",
			[]string{"hello", "world"},
		},
		{
			"Go, unlike many other languages, is really cool!",
			[]string{"go", "unlike", "many", "other", "languages", "is", "really", "cool"},
		},
		{
			"  There are     a bunch of   , extra spaces and punctuation  signs in this title    !  ",
			[]string{"there", "are", "a", "bunch", "of", "extra", "spaces", "and", "punctuation", "signs", "in", "this", "title"},
		},
	}

	for _, test := range tests {
		result := extractWords(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("extractWords(%q) = %v, but expected %v", test.input, result, test.expected)
		}
	}
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
