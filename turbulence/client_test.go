package turbulence_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-test/turbulence"
)

const expectedPOSTRequest = `{
	"Tasks": [{
		"Type": "kill"
	}],
	"Deployments": [{
		"Name": "deployment-name",
		"Jobs": [{
			"Name": "job-name",
			"Indices": [0]
		}]
	}]
}`

const POSTResponse = `{
	"ID": "someID",
	"ExecutionStartedAt": "0001-01-01T00:00:00Z",
	"ExecutionCompletedAt": "",
	"Events": null
}`

const GETResponse = `{
	"ID": "someID",
	"ExecutionStartedAt": "0001-01-01T00:00:00Z",
	"ExecutionCompletedAt": "0001-01-01T00:01:00Z",
	"Events": [
		{"Error": ""}
	]
}`

var _ = Describe("Client", func() {
	Describe("KillIndices", func() {
		It("Makes a POST request to create an incident", func() {
			var receivedPOSTBody []byte
			var errorReadingBody error

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch {
				case request.URL.Path == "/api/v1/incidents" && request.Method == "POST":
					receivedPOSTBody, errorReadingBody = ioutil.ReadAll(request.Body)
					defer request.Body.Close()
					writer.Write([]byte(POSTResponse))

				case request.URL.Path == "/api/v1/incidents/someID" && request.Method == "GET":
					writer.Write([]byte(GETResponse))
				}
			}))

			client := turbulence.NewClient(server.URL)
			errorKillingIndices := client.KillIndices("deployment-name", "job-name", []int{0})

			Expect(errorReadingBody).NotTo(HaveOccurred())
			Expect(errorKillingIndices).NotTo(HaveOccurred())
			Expect(string(receivedPOSTBody)).To(MatchJSON(expectedPOSTRequest))
		})
	})
})
