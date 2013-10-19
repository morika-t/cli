package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

func TestCreateUserProvidedServiceInstance(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/user_provided_service_instances",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":""}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, request)
	defer ts.Close()

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	apiResponse := repo.Create("my-custom-service", params, "")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestCreateUserProvidedServiceInstanceWithSyslogDrain(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/user_provided_service_instances",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":"syslog://example.com"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, request)
	defer ts.Close()

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	apiResponse := repo.Create("my-custom-service", params, "syslog://example.com")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateUserProvidedServiceInstance(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/user_provided_service_instances/my-instance-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"}}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, request)
	defer ts.Close()

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	apiResponse := repo.Update(cf.ServiceInstance{Guid: "my-instance-guid"}, params)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createUserProvidedServiceInstanceRepo(t *testing.T, request testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo UserProvidedServiceInstanceRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{request})

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Space:       cf.Space{Guid: "my-space-guid"},
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCCUserProvidedServiceInstanceRepository(config, gateway)
	return
}
