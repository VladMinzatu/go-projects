package ports

import "github.com/VladMinzatu/go-projects/hn-scan/core/domain"

type TopStoriesRepo interface {
	GetTopStories(n int) ([]domain.Story, error)
}
