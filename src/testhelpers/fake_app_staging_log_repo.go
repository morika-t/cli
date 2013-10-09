package testhelpers

import (
	"cf/net"
	"io"
)

type FakeStagingLogRepo struct{
	StreamLogUrl string
	StreamLogResponse io.Reader
}


func (repo FakeStagingLogRepo) StreamLog(logUrl string) (logStream io.Reader, apiResponse net.ApiResponse) {
	repo.StreamLogUrl = logUrl
	logStream = repo.StreamLogResponse

	return
}
