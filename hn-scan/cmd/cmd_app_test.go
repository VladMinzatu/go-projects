package cmd

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/VladMinzatu/go-projects/hn-scan/core/domain"
	"github.com/VladMinzatu/go-projects/hn-scan/core/service"
)

func TestCmdRun(t *testing.T) {
	tests := []struct {
		args          []string
		service       HNServiceMock
		expectedCall  bool
		expectedError error
		expectedLimit int
		expectedTerms []string
	}{
		{
			args:          []string{},
			service:       HNServiceMock{},
			expectedCall:  true,
			expectedError: nil,
			expectedLimit: defaultStories,
			expectedTerms: []string{},
		},
		{
			args:          []string{"-h"},
			service:       HNServiceMock{},
			expectedCall:  false,
			expectedError: errors.New("flag: help requested"),
			expectedLimit: 0,
			expectedTerms: nil,
		},
		{
			args:          []string{"-n", "10"},
			service:       HNServiceMock{},
			expectedCall:  true,
			expectedError: nil,
			expectedLimit: 10,
			expectedTerms: []string{},
		},
		{
			args:          []string{"-n", "200"},
			service:       HNServiceMock{},
			expectedCall:  false,
			expectedError: errors.New("Requested number of stories out of bounds [1, 100]"),
			expectedLimit: 0,
			expectedTerms: nil,
		},
		{
			args:          []string{"-term", "go"},
			service:       HNServiceMock{},
			expectedCall:  true,
			expectedError: nil,
			expectedLimit: defaultStories,
			expectedTerms: []string{"go"},
		},
		{
			args:          []string{"-term", "go", "-term", "build"},
			service:       HNServiceMock{},
			expectedCall:  true,
			expectedError: nil,
			expectedLimit: defaultStories,
			expectedTerms: []string{"go", "build"},
		},
		{
			args:          []string{"-n", "10", "-term", "go"},
			service:       HNServiceMock{},
			expectedCall:  true,
			expectedError: nil,
			expectedLimit: 10,
			expectedTerms: []string{"go"},
		},
		{
			args:          []string{"-n", "10", "-term", "go", "-term", "build"},
			service:       HNServiceMock{},
			expectedCall:  true,
			expectedError: nil,
			expectedLimit: 10,
			expectedTerms: []string{"go", "build"},
		},
		{
			args:          []string{"-term", "go", "-n", "10", "-term", "build"},
			service:       HNServiceMock{},
			expectedCall:  true,
			expectedError: nil,
			expectedLimit: 10,
			expectedTerms: []string{"go", "build"},
		},
		{
			args:          []string{"-n", "10", "-term", "go", "-x", "10"},
			service:       HNServiceMock{},
			expectedCall:  false,
			expectedError: errors.New("flag provided but not defined: -x"),
			expectedLimit: 0,
			expectedTerms: nil,
		},
		{
			args:          []string{"-n", "10", "-term", "go", "foobar"},
			service:       HNServiceMock{},
			expectedCall:  false,
			expectedError: errors.New("Positional arguments specified"),
			expectedLimit: 0,
			expectedTerms: nil,
		},
	}

	byteBuf := new(bytes.Buffer)
	for _, tc := range tests {
		app := CmdApp{&tc.service}
		err := app.Run(byteBuf, tc.args)
		if tc.expectedError == nil && err != nil {
			t.Errorf("Expected nil error, got: %v\n", err)
		}
		if tc.expectedError != nil && err.Error() != tc.expectedError.Error() {
			t.Errorf("Expected error to be: %v, got: %v\n", tc.expectedError, err)
		}

		if !tc.expectedCall {
			continue
		} else {
			if tc.service.requestSpy == nil {
				t.Error("Expected call to service, but none recorded")
				continue
			}
		}
		if tc.service.requestSpy.Limit() != tc.expectedLimit {
			t.Errorf("Expected limit to be: %v, got: %v\n", tc.expectedLimit, tc.service.requestSpy.Limit())
		}

		if !reflect.DeepEqual(tc.service.requestSpy.Terms(), tc.expectedTerms) {
			t.Errorf("Expected terms to be: %v, got: %v\n", tc.expectedTerms, tc.service.requestSpy.Terms())
		}
		byteBuf.Reset()
	}
}

type HNServiceMock struct {
	stories    []domain.Story
	err        error
	requestSpy *service.HNServiceRequest
}

func (service *HNServiceMock) GetTopStories(request *service.HNServiceRequest) ([]domain.Story, error) {
	service.requestSpy = request
	if service.err != nil {
		return nil, service.err
	}
	return service.stories, nil
}
