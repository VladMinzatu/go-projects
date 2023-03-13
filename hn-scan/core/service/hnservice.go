package service

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
	"github.com/VladMinzatu/go-projects/hn-scan/core/ports"
)

const (
	minStories = 1
	maxStories = 100
)

type HNService struct {
	topStoriesRepo ports.TopStoriesRepo
}

func NewHNService(topStoriesRepo ports.TopStoriesRepo) HNService {
	return HNService{topStoriesRepo: topStoriesRepo}
}

func (service HNService) GetTopStories(request *HNServiceRequest) ([]domain.Story, error) {
	stories, err := service.topStoriesRepo.GetTopStories(request.Limit())
	if err != nil {
		return nil, err
	}

	if len(request.Terms()) == 0 {
		return stories, nil
	}

	var result []domain.Story
	termsMap := make(map[string]bool)
	for _, term := range request.Terms() {
		termsMap[strings.ToLower(term)] = true
	}

	for _, story := range stories {
		words := extractWords(story.Title)
		for _, word := range words {
			if termsMap[word] {
				result = append(result, story)
				break
			}
		}
	}

	return result, nil
}

type HNServiceRequest struct {
	limit int
	terms []string
}

func NewHNServiceRequest(limit int, terms []string) (*HNServiceRequest, error) {
	if limit < minStories || limit > maxStories {
		return nil, fmt.Errorf("Requested number of stories out of bounds [%d, %d]", minStories, maxStories)
	}
	return &HNServiceRequest{limit: limit, terms: terms}, nil
}

func (req *HNServiceRequest) Limit() int {
	return req.limit
}

func (req *HNServiceRequest) Terms() []string {
	return req.terms
}

func extractWords(text string) []string {
	var words []string
	separatorFunc := func(c rune) bool {
		return unicode.IsSpace(c) || unicode.IsPunct(c)
	}

	for _, word := range strings.FieldsFunc(text, separatorFunc) {
		if len(word) > 0 {
			words = append(words, strings.ToLower(word))
		}
	}
	return words
}
