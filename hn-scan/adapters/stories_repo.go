package adapters

import (
	"sort"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
)

type TopStoriesRepo struct {
	client HackerNewsClient
}

type HackerNewsClient interface {
	GetTopStoryIds() ([]int, error)
	ResolveStory(id int) (domain.Story, error)
}

func NewTopStoriesRepo(client HackerNewsClient) *TopStoriesRepo {
	return &TopStoriesRepo{client: client}
}

func (repo *TopStoriesRepo) GetTopStories(n int) ([]domain.Story, error) {
	storyIds, err := repo.getStoryIds(n)
	if err != nil {
		return nil, err
	}
	return repo.resolveStories(storyIds)
}

func (repo *TopStoriesRepo) getStoryIds(n int) ([]int, error) {
	storyIds, err := repo.client.GetTopStoryIds()
	if err != nil {
		return storyIds, err
	}

	return storyIds[:n], nil
}

type Result struct {
	idx   int
	story domain.Story
	err   error
}

func (repo *TopStoriesRepo) resolveStories(ids []int) ([]domain.Story, error) {
	ch := make(chan Result)
	for idx, id := range ids {
		storyId := id // TODO: needed for concurrently fetching all ids and not repeating the same id??
		go repo.resolveStory(idx, storyId, ch)
	}
	var results []Result
	for range ids {
		result := <-ch
		if result.err != nil {
			// TODO: log warn in fetching
		} else {
			results = append(results, result)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].idx < results[j].idx
	})

	var stories []domain.Story
	for _, result := range results {
		stories = append(stories, domain.Story{Title: result.story.Title, Url: result.story.Url})
	}

	return stories, nil
}

func (repo *TopStoriesRepo) resolveStory(idx, id int, ch chan<- Result) (domain.Story, error) {
	story, err := repo.client.ResolveStory(id)
	ch <- Result{idx: idx, story: story, err: err}
	return story, nil
}
