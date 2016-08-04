package bosh_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveManifestVersions", func() {
	It("resolves the latest versions of releases", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			switch r.URL.Path {
			case "/releases/consats":
				Expect(r.Method).To(Equal("GET"))
				w.Write([]byte(`{"versions":["2.0.0","3.0.0","4.0.0"]}`))
			case "/stemcells":
				Expect(r.Method).To(Equal("GET"))
				w.Write([]byte(`[
					{"name": "some-stemcell-name","version": "1.0.0"},
					{"name": "some-stemcell-name","version": "2.0.0"},
					{"name": "some-other-stemcell-name","version": "100.0.0"}
				]`))
			default:
				Fail("unexpected route")
			}
		}))

		client := bosh.NewClient(bosh.Config{
			URL:                 server.URL,
			Username:            "some-username",
			Password:            "some-password",
			TaskPollingInterval: time.Nanosecond,
		})

		manifest := `---
director_uuid: some-director-uuid
name: some-name
compilation: some-compilation-value
update: some-update-value
networks: some-networks-value
resource_pools:
- name: some-resource-pool-1
  network: some-network
  size: some-size
  cloud_properties: some-cloud-properties
  env: some-env
  stemcell:
    name: "some-stemcell-name"
    version: 1.0.0
- name: some-resource-pool-2
  network: some-network
  stemcell:
    name: "some-stemcell-name"
    version: latest
- name: some-resource-pool-3
  network: some-network
  stemcell:
    name: "some-other-stemcell-name"
    version: latest
jobs: some-jobs-value
properties: some-properties-value
releases:
- name: consul
  version: 2.0.0
- name: consats
  version: latest
`

		resolvedManifest, err := client.ResolveManifestVersions([]byte(manifest))
		Expect(err).NotTo(HaveOccurred())
		Expect(resolvedManifest).To(gomegamatchers.MatchYAML(`---
director_uuid: some-director-uuid
name: some-name
compilation: some-compilation-value
update: some-update-value
networks: some-networks-value
resource_pools:
- name: some-resource-pool-1
  network: some-network
  size: some-size
  cloud_properties: some-cloud-properties
  env: some-env
  stemcell:
    name: "some-stemcell-name"
    version: 1.0.0
- name: some-resource-pool-2
  network: some-network
  stemcell:
    name: "some-stemcell-name"
    version: 2.0.0
- name: some-resource-pool-3
  network: some-network
  stemcell:
    name: "some-other-stemcell-name"
    version: 100.0.0
jobs: some-jobs-value
properties: some-properties-value
releases:
- name: consul
  version: 2.0.0
- name: consats
  version: 4.0.0
`))
	})

	Context("failure cases", func() {
		Context("when the yaml is malformed", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{})
				_, err := client.ResolveManifestVersions([]byte("%%%"))
				Expect(err).To(MatchError(ContainSubstring("yaml: ")))
			})
		})

		Context("when the stemcell API causes an error", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))
					w.WriteHeader(http.StatusNotFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})
				manifest := `---
resource_pools:
- name: some-resource-pool
  network: some-network
  stemcell:
    name: "some-other-stemcell-name"
    version: latest
`

				_, err := client.ResolveManifestVersions([]byte(manifest))
				Expect(err).To(MatchError("stemcell some-other-stemcell-name could not be found"))
			})
		})

		Context("when the stemcell cannot resolve the latest", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))
					Expect(r.Method).To(Equal("GET"))

					username, password, ok := r.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(username).To(Equal("some-username"))
					Expect(password).To(Equal("some-password"))

					w.Write([]byte(`[]`))

				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})
				manifest := `---
resource_pools:
- name: some-resource-pool
  network: some-network
  stemcell:
    name: "some-other-stemcell-name"
    version: latest
`

				_, err := client.ResolveManifestVersions([]byte(manifest))
				Expect(err).To(MatchError("no stemcell versions found, cannot get latest"))
			})
		})

		Context("when the release API causes an error", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/releases/some-release-name"))
					w.WriteHeader(http.StatusNotFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})
				manifest := `---
releases:
- name: consul
  version: 2.0.0
- name: some-release-name
  version: latest
`

				_, err := client.ResolveManifestVersions([]byte(manifest))
				Expect(err).To(MatchError("release some-release-name could not be found"))
			})
		})
	})
})
