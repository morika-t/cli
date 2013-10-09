package api

import (
	"cf/configuration"
	"cf/net"
	"io"
)

type AppStagingLogRepository interface {
	StreamLog(logUrl string) (logStream io.Reader, apiResponse net.ApiResponse)
}

type CloudControllerAppStagingLogRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerAppStagingLogRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerAppStagingLogRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppStagingLogRepository) StreamLog(logUrl string) (logStream io.Reader, apiResponse net.ApiResponse) {
	request, apiResponse := repo.gateway.NewRequest("GET", logUrl, "", nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	logStream, _, apiResponse = repo.gateway.PerformRequestForResponseByteStream(request)

	return
}
