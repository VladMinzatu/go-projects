package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/VladMinzatu/go-projects/hn-scan/adapters"
	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
	"github.com/VladMinzatu/go-projects/hn-scan/core/service"
)

const (
	defaultStories = 50
)

type termsParam []string

func (t *termsParam) String() string {
	return fmt.Sprintf("%s", *t)
}

func (t *termsParam) Set(value string) error {
	*t = append(*t, value)
	return nil
}

type config struct {
	numStories int
	terms      termsParam
	debug      bool
}

type HNService interface {
	GetTopStories(request *service.HNServiceRequest) ([]domain.Story, error)
}

type CmdApp struct {
	svc HNService
}

func NewCmdApp() CmdApp {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	topStoriesUrl := "https://hacker-news.firebaseio.com/v0/topstories.json"
	storyResolutionUrl := "https://hacker-news.firebaseio.com/v0/item/"
	hnClient := adapters.NewHackerNewsClient(&httpClient, topStoriesUrl, storyResolutionUrl)
	service := service.NewHNService(adapters.NewTopStoriesRepo(hnClient))
	return CmdApp{svc: service}
}

func (app CmdApp) Run(w io.Writer, args []string) error {
	conf, err := parseArgs(w, args)
	if err != nil {
		return err
	}

	return app.run(w, conf)
}

func parseArgs(w io.Writer, args []string) (config, error) {
	c := config{numStories: defaultStories, terms: []string{}, debug: false}
	fs := flag.NewFlagSet("hn-scan-app", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.IntVar(&c.numStories, "n", defaultStories, "Number of stories to scan")
	fs.Var(&c.terms, "term", "Term to use for filtering stories.")
	fs.BoolVar(&c.debug, "debug", false, "Print debug logs")
	err := fs.Parse(args)
	if err != nil {
		return c, err
	}
	if fs.NArg() != 0 {
		return c, errors.New("Positional arguments specified")
	}
	return c, nil
}

func (app CmdApp) run(w io.Writer, config config) error {
	if config.debug {
		log.SetLevel(log.DebugLevel)
	}

	request, err := service.NewHNServiceRequest(config.numStories, config.terms)
	if err != nil {
		return err
	}
	result, err := app.svc.GetTopStories(request)
	if err != nil {
		return err
	}
	if len(result) == 0 {
		fmt.Fprintf(w, "No stories found in the top %d matching terms %q\n", config.numStories, config.terms)
		return nil
	}

	fmt.Fprintf(w, "Retrieved the following stories from the top %d matching terms %q:\n\n", config.numStories, config.terms)
	for _, story := range result {
		fmt.Fprintf(w, "%s (%s)\n", story.Title, story.Url)
	}
	return nil
}
