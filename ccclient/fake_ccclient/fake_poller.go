// This file was generated by counterfeiter
package fake_ccclient

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/cloudfoundry-incubator/file-server/ccclient"
)

type FakePoller struct {
	PollStub        func(fallbackURL *url.URL, res *http.Response, closeChan <-chan bool) error
	pollMutex       sync.RWMutex
	pollArgsForCall []struct {
		fallbackURL *url.URL
		res         *http.Response
		closeChan   <-chan bool
	}
	pollReturns struct {
		result1 error
	}
}

func (fake *FakePoller) Poll(fallbackURL *url.URL, res *http.Response, closeChan <-chan bool) error {
	fake.pollMutex.Lock()
	fake.pollArgsForCall = append(fake.pollArgsForCall, struct {
		fallbackURL *url.URL
		res         *http.Response
		closeChan   <-chan bool
	}{fallbackURL, res, closeChan})
	fake.pollMutex.Unlock()
	if fake.PollStub != nil {
		return fake.PollStub(fallbackURL, res, closeChan)
	} else {
		return fake.pollReturns.result1
	}
}

func (fake *FakePoller) PollCallCount() int {
	fake.pollMutex.RLock()
	defer fake.pollMutex.RUnlock()
	return len(fake.pollArgsForCall)
}

func (fake *FakePoller) PollArgsForCall(i int) (*url.URL, *http.Response, <-chan bool) {
	fake.pollMutex.RLock()
	defer fake.pollMutex.RUnlock()
	return fake.pollArgsForCall[i].fallbackURL, fake.pollArgsForCall[i].res, fake.pollArgsForCall[i].closeChan
}

func (fake *FakePoller) PollReturns(result1 error) {
	fake.PollStub = nil
	fake.pollReturns = struct {
		result1 error
	}{result1}
}

var _ ccclient.Poller = new(FakePoller)
