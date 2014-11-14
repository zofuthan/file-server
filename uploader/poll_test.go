package uploader_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudfoundry-incubator/file-server/uploader"
	"github.com/cloudfoundry-incubator/file-server/uploader/test_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Uploader", func() {
	var (
		u               uploader.Uploader
		transport       http.RoundTripper
		pollRequestChan chan *http.Request
	)

	Describe("Poll", func() {
		var (
			pollErrChan chan error

			pollURL                *url.URL
			originalUploadResponse *http.Response
			closeChan              chan bool
			pollInterval           time.Duration
		)

		BeforeEach(func() {
			closeChan = make(chan bool, 1)
			pollInterval = 10 * time.Millisecond
		})

		JustBeforeEach(func() {
			u = uploader.New(&url.URL{}, transport)
			pollErrChan = make(chan error, 1)
			go func(pec chan error) {
				defer GinkgoRecover()
				pec <- u.Poll(pollURL, originalUploadResponse, closeChan, pollInterval)
			}(pollErrChan)
		})

		Context("when the initial response is invalid", func() {
			BeforeEach(func() {
				originalUploadResponse = responseWithBody("garbage")
			})

			It("returns an error", func() {
				Eventually(pollErrChan).Should(Receive(HaveOccurred()))
			})
		})

		Context("with a valid initial response", func() {
			Context("when the status is 'finished'", func() {
				BeforeEach(func() {
					originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", uploader.JOB_FINISHED))
				})

				It("returns with no error", func() {
					Eventually(pollErrChan).Should(Receive(BeNil()))
				})
			})

			Context("when the status is 'failed'", func() {
				BeforeEach(func() {
					originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", uploader.JOB_FAILED))
				})

				It("returns with an error", func() {
					Eventually(pollErrChan).Should(Receive(MatchError("upload job failed")))
				})
			})

			Context("when the status is unrecognizable", func() {
				BeforeEach(func() {
					originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", "made-up-job-status"))
				})

				It("returns with an error", func() {
					Eventually(pollErrChan).Should(Receive(MatchError("unknown job status: made-up-job-status")))
				})
			})

			Context("when the status is 'queued'", func() {
				var jobStatus string

				BeforeEach(func() {
					jobStatus = uploader.JOB_QUEUED
				})

				Context("when the close channel is written to", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", jobStatus))
						closeChan <- true
					})

					It("errors", func() {
						Eventually(pollErrChan).Should(Receive(MatchError("upstream request was cancelled")))
					})
				})

				Context("when the metadata URL can't be parsed", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("http://%x", jobStatus))
					})

					It("errors", func() {
						Eventually(pollErrChan).Should(Receive(BeAssignableToTypeOf(&url.Error{})))
					})
				})

				Context("when the metadata URL includes a host", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", jobStatus))

						pollRequestChan = make(chan *http.Request, 5)
						transport = test_helpers.NewFakeRoundTripper(
							pollRequestChan,
							map[string]test_helpers.RespErrorPair{
								"example.com":          {responseWithBody(pollingResponseBody("http://polling-endpoint.com", uploader.JOB_QUEUED)), nil},
								"polling-endpoint.com": {responseWithBody(pollingResponseBody("http://2nd-time.com", uploader.JOB_FAILED)), nil},
							},
						)
					})

					It("uses the metadata URL to make the request to the polling endpoint", func() {
						var firstPollRequest *http.Request
						Eventually(pollRequestChan).Should(Receive(&firstPollRequest))
						Ω(firstPollRequest.URL.Host).Should(Equal("example.com"))

						var secondPollRequest *http.Request
						Eventually(pollRequestChan).Should(Receive(&secondPollRequest))
						Ω(secondPollRequest.URL.Host).Should(Equal("polling-endpoint.com"))
					})
				})

				Context("when the metadata URL is just a path", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("/just/a/path", jobStatus))

						pollRequestChan = make(chan *http.Request, 5)
						transport = test_helpers.NewFakeRoundTripper(
							pollRequestChan,
							map[string]test_helpers.RespErrorPair{
								"fallback-url.com":     {responseWithBody(pollingResponseBody("http://polling-endpoint.com", uploader.JOB_QUEUED)), nil},
								"polling-endpoint.com": {responseWithBody(pollingResponseBody("http://2nd-time.com", uploader.JOB_FAILED)), nil},
							},
						)

						pollURL, _ = url.Parse("http://fallback-url.com")
					})

					It("appends the host and scheme from the given poll URL to the new path to perform the request to the polling endpoint", func() {
						var firstPollRequest *http.Request
						Eventually(pollRequestChan).Should(Receive(&firstPollRequest))
						Ω(firstPollRequest.URL.Host).Should(Equal("fallback-url.com"))

						var secondPollRequest *http.Request
						Eventually(pollRequestChan).Should(Receive(&secondPollRequest))
						Ω(secondPollRequest.URL.Host).Should(Equal("polling-endpoint.com"))
					})
				})

				Context("when there is an error making a request to the polling endpoint", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", jobStatus))

						pollRequestChan = make(chan *http.Request, 5)
						transport = test_helpers.NewFakeRoundTripper(
							pollRequestChan,
							map[string]test_helpers.RespErrorPair{
								"example.com": {nil, errors.New("something bad")},
							},
						)
					})

					It("errors", func() {
						var urlErr error
						Eventually(pollErrChan).Should(Receive(&urlErr))
						Ω(urlErr).Should(MatchError(ContainSubstring("something bad")))
					})
				})

				Context("when the response from the polling endpoint is invalid", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", jobStatus))

						pollRequestChan = make(chan *http.Request, 5)
						transport = test_helpers.NewFakeRoundTripper(
							pollRequestChan,
							map[string]test_helpers.RespErrorPair{
								"example.com": {responseWithBody("garbage"), nil},
							},
						)
					})

					It("errors", func() {
						Eventually(pollErrChan).Should(Receive(BeAssignableToTypeOf(&json.SyntaxError{})))
					})
				})

				Context("when the responses from the polling endpoint eventually have a problem", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", jobStatus))

						pollRequestChan = make(chan *http.Request, 5)
						transport = test_helpers.NewFakeRoundTripper(
							pollRequestChan,
							map[string]test_helpers.RespErrorPair{
								"example.com": {responseWithBody(pollingResponseBody("http://1.com", uploader.JOB_QUEUED)), nil},
								"1.com":       {responseWithBody(pollingResponseBody("http://2.com", uploader.JOB_RUNNING)), nil},
								"2.com":       {responseWithBody(pollingResponseBody("http://3.com", uploader.JOB_RUNNING)), nil},
								"3.com":       {responseWithBody(pollingResponseBody("http://4.com", uploader.JOB_RUNNING)), nil},
								"4.com":       {responseWithBody("garbage"), nil},
							},
						)
					})

					It("eventually errors", func() {
						Eventually(pollErrChan).Should(Receive(HaveOccurred()))
					})
				})

				Context("when the responses from the polling endpoint eventually report a finished job", func() {
					BeforeEach(func() {
						originalUploadResponse = responseWithBody(pollingResponseBody("http://example.com", jobStatus))

						pollRequestChan = make(chan *http.Request, 5)
						transport = test_helpers.NewFakeRoundTripper(
							pollRequestChan,
							map[string]test_helpers.RespErrorPair{
								"example.com": {responseWithBody(pollingResponseBody("http://1.com", uploader.JOB_QUEUED)), nil},
								"1.com":       {responseWithBody(pollingResponseBody("http://2.com", uploader.JOB_RUNNING)), nil},
								"2.com":       {responseWithBody(pollingResponseBody("http://3.com", uploader.JOB_RUNNING)), nil},
								"3.com":       {responseWithBody(pollingResponseBody("http://4.com", uploader.JOB_RUNNING)), nil},
								"4.com":       {responseWithBody(pollingResponseBody("http://4.com", uploader.JOB_FINISHED)), nil},
							},
						)
					})

					It("returns with no error", func() {
						Eventually(pollErrChan).Should(Receive(BeNil()))
					})
				})
			})
		})
	})
})

func responseWithBody(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
	}
}

func pollingResponseBody(url, status string) string {
	return `{"metadata":{"url":"` + url + `"},"entity":{"status":"` + status + `"}}`
}