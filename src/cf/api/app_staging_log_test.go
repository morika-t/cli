package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var expectedResponse = `
log line 1
log line 2
log line 3
`

var appLogEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	methodMatches := request.Method == "GET"
	pathMatches := strings.Contains(request.URL.Path, "/path/to/logs")

	if !methodMatches || !pathMatches {
		fmt.Printf("One of the matchers did not match. Method [%t] Path [%t]",
			methodMatches, pathMatches)

		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	io.WriteString(writer, expectedResponse)
}

func TestStreamLog(t *testing.T) {
	appLogServer := httptest.NewTLSServer(http.HandlerFunc(appLogEndpoint))
	defer appLogServer.Close()

	config := &configuration.Configuration{
		Target:      appLogServer.URL,
		AccessToken: "BEARER my_access_token",
	}

	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerAppStagingLogRepository(config, gateway)

	logReader, apiResponse := repo.StreamLog(appLogServer.URL + "/path/to/logs")
	assert.False(t, apiResponse.IsNotSuccessful())

	responseBytes, err := ioutil.ReadAll(logReader)
	assert.NoError(t, err)

	response := string(responseBytes)
	assert.Equal(t, response, expectedResponse)
}
