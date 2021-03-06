package core

import (
	"encoding/base64"
	"math/rand"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResponseWriter tests", func() {
	Context("writing to response object", func() {
		response := NewProxyResponseWriter()

		It("Sets the correct default status", func() {
			Expect(defaultStatusCode).To(Equal(response.status))
		})

		It("Initializes the headers map", func() {
			Expect(response.headers).ToNot(BeNil())
			Expect(0).To(Equal(len(response.headers)))
		})

		It("Writes headers correctly", func() {
			response.Header().Add("Content-Type", "application/json")

			Expect(1).To(Equal(len(response.headers)))
			Expect("application/json").To(Equal(response.headers["Content-Type"][0]))
		})

		It("Writes body content correctly", func() {
			binaryBody := make([]byte, 256)
			_, err := rand.Read(binaryBody)
			Expect(err).To(BeNil())

			written, err := response.Write(binaryBody)
			Expect(err).To(BeNil())
			Expect(len(binaryBody)).To(Equal(written))
		})

		It("Automatically set the status code to 200", func() {
			Expect(http.StatusOK).To(Equal(response.status))
		})

		It("Forces the status to a new code", func() {
			response.WriteHeader(http.StatusAccepted)
			Expect(http.StatusAccepted).To(Equal(response.status))
		})
	})

	Context("Export API Gateway proxy response", func() {
		noHeaderResponse := NewProxyResponseWriter()

		It("Refuses responses with no headers", func() {
			_, err := noHeaderResponse.GetProxyResponse()
			Expect(err).ToNot(BeNil())
			Expect("No headers generated for response").To(Equal(err.Error()))
		})

		emtpyResponse := NewProxyResponseWriter()
		emtpyResponse.Header().Add("Content-Type", "application/json")

		It("Refuses empty responses with default status code", func() {
			_, err := emtpyResponse.GetProxyResponse()
			Expect(err).ToNot(BeNil())
			Expect("Status code not set on response").To(Equal(err.Error()))
		})

		simpleResponse := NewProxyResponseWriter()
		simpleResponse.Write([]byte("hello"))
		simpleResponse.Header().Add("Content-Type", "text/plain")
		It("Writes text body correctly", func() {
			proxyResponse, err := simpleResponse.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(proxyResponse).ToNot(BeNil())

			Expect("hello").To(Equal(proxyResponse.Body))
			Expect(http.StatusOK).To(Equal(proxyResponse.StatusCode))
			Expect(1).To(Equal(len(proxyResponse.Headers)))
			Expect("text/plain").To(Equal(proxyResponse.Headers["Content-Type"]))
			Expect(proxyResponse.IsBase64Encoded).To(BeFalse())
		})

		binaryResponse := NewProxyResponseWriter()
		binaryResponse.Header().Add("Content-Type", "application/octet-stream")
		binaryBody := make([]byte, 256)
		_, err := rand.Read(binaryBody)
		if err != nil {
			Fail("Could not generate random binary body")
		}
		binaryResponse.Write(binaryBody)
		binaryResponse.WriteHeader(http.StatusAccepted)

		It("Encodes binary responses correctly", func() {
			proxyResponse, err := binaryResponse.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(proxyResponse).ToNot(BeNil())

			Expect(proxyResponse.IsBase64Encoded).To(BeTrue())
			Expect(base64.StdEncoding.EncodedLen(len(binaryBody))).To(Equal(len(proxyResponse.Body)))

			Expect(base64.StdEncoding.EncodeToString(binaryBody)).To(Equal(proxyResponse.Body))
			Expect(1).To(Equal(len(proxyResponse.Headers)))
			Expect("application/octet-stream").To(Equal(proxyResponse.Headers["Content-Type"]))
			Expect(http.StatusAccepted).To(Equal(proxyResponse.StatusCode))
		})
	})
})
