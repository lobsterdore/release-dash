package testintegration

import (
	"context"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var intSupport *integrationSupport

var _ = Describe("Release Dash", func() {

	BeforeSuite(func() {
		var err error
		intSupport, err = NewIntegrationSupport()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		err := intSupport.teardown()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Test Homepage", func() {
		It("shows changelogs", func() {
			ctx := context.Background()

			serviceEndpoint, err := intSupport.AppContainer.Endpoint(ctx, "")
			Expect(err).ToNot(HaveOccurred())

			client := &http.Client{}
			req, err := http.NewRequest("GET", "http://"+serviceEndpoint+"/", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			bodyRaw, _ := ioutil.ReadAll(resp.Body)
			body := string(bodyRaw)

			Expect(body).Should(ContainSubstring("test-repo"))
			Expect(body).Should(ContainSubstring("from-tag > to-tag"))
			Expect(body).Should(ContainSubstring("test-commit"))
		})

	})
})
