package bosh_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadRelease", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/releases":
				Expect(req.Method).To(Equal("POST"))
				username, password, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				contents, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(Equal([]byte("I am a banana!")))

				w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/2", req.Host))
				w.WriteHeader(http.StatusFound)

			case "/tasks/2":
				Expect(req.Method).To(Equal("GET"))
				username, password, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				w.Write([]byte(`{"id": 2, "state": "done"}`))

			default:
				Fail(fmt.Sprintf("unhandled request to %s", req.URL.Path))
			}
		}))
	})

	It("uploads the release to the director", func() {
		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		taskID, err := client.UploadRelease(strings.NewReader("I am a banana!"))
		Expect(err).NotTo(HaveOccurred())
		Expect(taskID).To(Equal(2))
	})

	Context("failure cases", func() {
		Context("when the request cannot be created", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "%%%%%",
				})

				_, err := client.UploadRelease(strings.NewReader("I am a banana!"))
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		Context("when the request cannot be made", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "",
				})

				_, err := client.UploadRelease(strings.NewReader("I am a banana!"))
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
			})
		})

		Context("when the request returns an unexpected response status", func() {
			It("returns an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}))

				client := bosh.NewClient(bosh.Config{
					URL: server.URL,
				})

				_, err := client.UploadRelease(strings.NewReader("Hi"))
				Expect(err).To(MatchError("unexpected response 418 I'm a teapot"))
			})
		})
	})
})
