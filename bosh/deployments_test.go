package bosh_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployments", func() {
	It("retrieves all deployments from the director", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/deployments"))
			Expect(r.Method).To(Equal("GET"))

			username, password, ok := r.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			w.Write([]byte(`[
					{"name": "deployment1"},
					{"name": "deployment2"},
					{"name": "deployment3"}
				]`))
		}))

		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		deployments, err := client.Deployments()
		Expect(err).NotTo(HaveOccurred())
		Expect(deployments).To(Equal(
			[]bosh.Deployment{
				{
					Name: "deployment1",
				},
				{
					Name: "deployment2",
				},
				{
					Name: "deployment3",
				},
			},
		))
	})

	Context("failure cases", func() {
		It("error on a malformed URL", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      "%%%%%%%%",
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("error on an empty URL", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      "",
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("errors on an unexpected status code with a body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError("unexpected response 502 Bad Gateway:\nMore Info"))
		})

		It("error on malformed JSON", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`%%%%%%%%`))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})

		It("returns an error on a bogus response body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})

			_, err := client.Deployments()

			Expect(err).To(MatchError("a bad read happened"))
		})
	})
})
