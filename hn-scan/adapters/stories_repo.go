package adapters

import (
	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
)

type TopStoriesRepo struct {
	client HackerNewsClient
}

type HackerNewsClient interface {
	GetTopStoryIds() ([]string, error)
	ResolveStory(id string) (domain.Story, error)
}

func NewTopStoriesRepo(client HackerNewsClient) TopStoriesRepo {
	return TopStoriesRepo{client: client}
}

func (repo *TopStoriesRepo) GetTopStories(n int) ([]domain.Story, error) {
	storyIds, err := repo.getStoryIds(n)
	if err != nil {
		return nil, err
	}
	return repo.resolveStories(storyIds)
}

func (repo *TopStoriesRepo) getStoryIds(n int) ([]string, error) {
	storyIds, err := repo.client.GetTopStoryIds()
	if err != nil {
		return storyIds, err
	}

	return storyIds, nil
}

type Result struct {
	story domain.Story
	err   error
}

func (repo *TopStoriesRepo) resolveStories(ids []string) ([]domain.Story, error) {
	ch := make(chan Result)
	for _, id := range ids {
		storyId := id // TODO: needed for concurrently fetching all ids and not repeating the same id??
		go repo.resolveStory(storyId, ch)
	}
	var stories []domain.Story
	for range ids {
		result := <-ch
		if result.err != nil {
			// TODO: log warn in fetching
		}
		stories = append(stories, domain.Story{Title: result.story.Title, Url: result.story.Url})
	}
	return stories, nil
}

func (repo *TopStoriesRepo) resolveStory(id string, ch chan<- Result) (domain.Story, error) {
	story, err := repo.client.ResolveStory(id)
	ch <- Result{story: story, err: err}
	return story, nil
}
