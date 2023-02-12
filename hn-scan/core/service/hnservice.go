package service

import (
	"fmt"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
	"github.com/VladMinzatu/go-projects/hn-scan/core/ports"
)

const (
	minStories = 1
	maxStories = 50
)

var ErrInvalidInput = fmt.Errorf("Requested number of stories out of bounds [%d, %d]", minStories, maxStories)

type HNService struct {
	topStoriesRepo ports.TopStoriesRepo
}

func NewHNService(topStoriesRepo ports.TopStoriesRepo) HNService {
	return HNService{topStoriesRepo: topStoriesRepo}
}

func (service HNService) GetTopStories(n int) ([]domain.Story, error) {
	if n < minStories || n > maxStories {
		return nil, ErrInvalidInput
	}

	//TODO: wrap the error?
	return service.topStoriesRepo.GetTopStories(n)
}
