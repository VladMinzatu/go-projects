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

type HNService struct {
	topStoriesRepo ports.TopStoriesRepo
}

func NewHNService(topStoriesRepo ports.TopStoriesRepo) HNService {
	return HNService{topStoriesRepo: topStoriesRepo}
}

func (service HNService) GetTopStories(request *HNServiceRequest) ([]domain.Story, error) {

	//TODO: wrap the error?
	return service.topStoriesRepo.GetTopStories(request.Limit())
}

type HNServiceRequest struct {
	limit int
	terms []string
}

func NewHNServiceRequest(limit int, terms []string) (*HNServiceRequest, error) {
	if limit < minStories || limit > maxStories {
		return nil, fmt.Errorf("Requested number of stories out of bounds [%d, %d]", minStories, maxStories)
	}
	if terms == nil {
		return &HNServiceRequest{limit: limit, terms: []string{}}, nil
	} else {
		return &HNServiceRequest{limit: limit, terms: terms}, nil
	}
}

func (req *HNServiceRequest) Limit() int {
	return req.limit
}

func (req *HNServiceRequest) Terms() []string {
	return req.terms
}
