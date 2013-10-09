package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"strings"
	"testhelpers"
	"testing"
)

var defaultAppForStart = cf.Application{
	Name:      "my-app",
	Guid:      "my-app-guid",
	Instances: 2,
	Urls:      []string{"http://my-app.example.com"},
}

var expectedStagingLog = "log line 1"

func createStagingLogRepo() testhelpers.FakeStagingLogRepo {
	return testhelpers.FakeStagingLogRepo{
		StreamLogResponse: strings.NewReader(expectedStagingLog),
	}
}

func readString(reader io.Reader) string {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func startAppWithInstancesAndErrors(app cf.Application, instances [][]cf.ApplicationInstance, errorCodes []string) (ui *testhelpers.FakeUI, appRepo *testhelpers.FakeApplicationRepository, stagingLogRepo testhelpers.FakeStagingLogRepo, reqFactory *testhelpers.FakeReqFactory) {
	config := &configuration.Configuration{ApplicationStartTimeout: 2}

	appRepo = &testhelpers.FakeApplicationRepository{
		FindByNameApp:          app,
		GetInstancesResponses:  instances,
		GetInstancesErrorCodes: errorCodes,
	}
	stagingLogRepo = createStagingLogRepo()
	args := []string{"my-app"}
	reqFactory = &testhelpers.FakeReqFactory{Application: app}
	ui = callStart(args, config, reqFactory, appRepo, stagingLogRepo)
	return
}

func TestStartCommandFailsWithUsage(t *testing.T) {
	config := &configuration.Configuration{}
	appRepo := &testhelpers.FakeApplicationRepository{
		GetInstancesResponses: [][]cf.ApplicationInstance{
			[]cf.ApplicationInstance{},
		},
		GetInstancesErrorCodes: []string{""},
	}
	reqFactory := &testhelpers.FakeReqFactory{}

	ui := callStart([]string{}, config, reqFactory, appRepo, createStagingLogRepo())
	assert.True(t, ui.FailedWithUsage)

	ui = callStart([]string{"my-app"}, config, reqFactory, appRepo, createStagingLogRepo())
	assert.False(t, ui.FailedWithUsage)
}

func TestStartApplication(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceRunning},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
	}

	errorCodes := []string{"", ""}
	ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], expectedStagingLog)
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "Started: app my-app available at http://my-app.example.com")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "my-app-guid")
}

func TestStartApplicationWhenAppHasNoURL(t *testing.T) {
	app := defaultAppForStart
	app.Urls = []string{}

	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceRunning},
		},
	}

	errorCodes := []string{""}
	ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(app, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], expectedStagingLog)
	assert.Contains(t, ui.Outputs[3], "Started")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "my-app-guid")
}

func TestStartApplicationWhenAppIsStillStaging(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{},
		[]cf.ApplicationInstance{},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceDown},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceRunning},
			cf.ApplicationInstance{State: cf.InstanceRunning},
		},
	}

	errorCodes := []string{api.APP_NOT_STAGED, api.APP_NOT_STAGED, "", "", ""}

	ui, _, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], expectedStagingLog)
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[4], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[5], "Started: app my-app available at http://my-app.example.com")
}

func TestStartApplicationWhenStagingFails(t *testing.T) {
	instances := [][]cf.ApplicationInstance{[]cf.ApplicationInstance{}}
	errorCodes := []string{"170001"}

	ui, _, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, instances, errorCodes)

	println(ui.DumpOutputs())
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], expectedStagingLog)
	assert.Contains(t, ui.Outputs[4], "FAILED")
	assert.Contains(t, ui.Outputs[5], "Error staging app")
}

func TestStartApplicationWhenOneInstanceFlaps(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceFlapping},
		},
	}

	errorCodes := []string{"", ""}

	ui, _, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], expectedStagingLog)
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "FAILED")
	assert.Contains(t, ui.Outputs[5], "Start unsuccessful")
}

func TestStartApplicationWhenStartTimesOut(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceDown},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceDown},
			cf.ApplicationInstance{State: cf.InstanceDown},
		},
	}

	errorCodes := []string{"", "", ""}

	ui, _, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], expectedStagingLog)
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[5], "0 of 2 instances running (2 down)")
	assert.Contains(t, ui.Outputs[6], "FAILED")
	assert.Contains(t, ui.Outputs[7], "Start app timeout")
}

func TestStartApplicationWhenStartFails(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{FindByNameApp: app, StartAppErr: true}
	stagingLogRepo := testhelpers.FakeStagingLogRepo{}
	args := []string{"my-app"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app}

	ui := callStart(args, config, reqFactory, appRepo, stagingLogRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Error starting application")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "my-app-guid")
}

func TestStartApplicationIsAlreadyStarted(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", State: "started"}
	appRepo := &testhelpers.FakeApplicationRepository{FindByNameApp: app}
	stagingLogRepo := testhelpers.FakeStagingLogRepo{}
	reqFactory := &testhelpers.FakeReqFactory{Application: app}

	args := []string{"my-app"}
	ui := callStart(args, config, reqFactory, appRepo, stagingLogRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already started")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "")
}

func callStart(args []string, config *configuration.Configuration, reqFactory *testhelpers.FakeReqFactory, appRepo api.ApplicationRepository, stagingLogRepo api.AppStagingLogRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("start", args)

	cmd := NewStart(ui, config, appRepo, stagingLogRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
