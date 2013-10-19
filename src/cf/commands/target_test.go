package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func getTargetDependencies() (orgRepo *testapi.FakeOrgRepository,
	spaceRepo *testapi.FakeSpaceRepository,
	configRepo *testconfig.FakeConfigRepository,
	reqFactory *testreq.FakeReqFactory) {

	orgRepo = &testapi.FakeOrgRepository{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	configRepo = &testconfig.FakeConfigRepository{}
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func TestTargetFailsWithUsage(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callTarget([]string{"foo"}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestTargetRequirements(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	reqFactory.LoginSuccess = true

	callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestTargetWithoutArgumentAndLoggedIn(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	config := configRepo.Login()
	config.Target = "https://api.run.pivotal.io"

	ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Outputs[0], "No org targeted")
	assert.Contains(t, ui.Outputs[1], "No space targeted")
}

// Start test with organization option
func TestTargetOrganizationWhenUserHasAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Login()
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.Space = cf.Space{Name: "my-space", Guid: "my-space-guid"}

	orgs := []cf.Organization{
		cf.Organization{Name: "my-organization", Guid: "my-organization-guid"},
	}
	orgRepo.Organizations = orgs
	orgRepo.FindByNameOrganization = orgs[0]

	ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, orgRepo.FindByNameName, "my-organization")
	assert.True(t, ui.ShowConfigurationCalled)

	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.Organization.Guid, "my-organization-guid")
}

func TestTargetOrganizationWhenUserDoesNotHaveAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	configRepo.Login()

	orgs := []cf.Organization{}
	orgRepo.Organizations = orgs
	orgRepo.FindByNameErr = true

	ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "No org targeted")

	ui = callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")

	ui = callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "No org targeted")
}

func TestTargetOrganizationWhenOrgNotFound(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	configRepo.Delete()
	configRepo.Login()

	config, err := configRepo.Get()
	assert.NoError(t, err)
	org := cf.Organization{Guid: "previous-org-guid", Name: "previous-org"}
	config.Organization = org
	err = configRepo.Save()
	assert.NoError(t, err)

	orgRepo.FindByNameNotFound = true

	ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "my-organization")
	assert.Contains(t, ui.Outputs[1], "not found")
}

// End test with organization option

// Start test with space option

func TestTargetSpaceWhenNoOrganizationIsSelected(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	configRepo.Login()

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "An org must be targeted before targeting a space")
	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.Organization.Guid, "")
}

func TestTargetSpaceWhenUserHasAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}

	spaces := []cf.Space{
		cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	spaceRepo.Spaces = spaces
	spaceRepo.FindByNameSpace = spaces[0]

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, spaceRepo.FindByNameName, "my-space")
	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.True(t, ui.ShowConfigurationCalled)
}

func TestTargetSpaceWhenUserDoesNotHaveAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}

	spaceRepo.FindByNameErr = true

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "my-space")

	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.Space.Guid, "")
	assert.True(t, ui.ShowConfigurationCalled)
}

func TestTargetSpaceWhenSpaceNotFound(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}

	spaceRepo.FindByNameNotFound = true

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "my-space")
	assert.Contains(t, ui.Outputs[1], "not found")
}

// End test with space option

// Targeting both org and space

func TestTargetOrganizationAndSpace(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	configRepo.Delete()
	configRepo.Login()

	org := cf.Organization{Name: "my-organization", Guid: "my-organization-guid"}
	orgRepo.FindByNameOrganization = org

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	spaceRepo.FindByNameSpace = space

	ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	savedConfig := testconfig.SavedConfiguration
	assert.True(t, ui.ShowConfigurationCalled)

	assert.Equal(t, orgRepo.FindByNameName, "my-organization")
	assert.Equal(t, savedConfig.Organization.Guid, "my-organization-guid")

	assert.Equal(t, spaceRepo.FindByNameName, "my-space")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
}

func TestTargetOrganizationAndSpaceWhenSpaceFails(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	configRepo.Delete()
	configRepo.Login()

	org := cf.Organization{Name: "my-organization", Guid: "my-organization-guid"}
	orgRepo.FindByNameOrganization = org

	spaceRepo.FindByNameErr = true

	ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	savedConfig := testconfig.SavedConfiguration
	assert.True(t, ui.ShowConfigurationCalled)

	assert.Equal(t, orgRepo.FindByNameName, "my-organization")
	assert.Equal(t, savedConfig.Organization.Guid, "my-organization-guid")
	assert.Equal(t, spaceRepo.FindByNameName, "my-space")
	assert.Equal(t, savedConfig.Space.Guid, "")
	assert.Contains(t, ui.Outputs[0], "FAILED")
}

// End test with org and space options

func callTarget(args []string,
	reqFactory *testreq.FakeReqFactory,
	configRepo configuration.ConfigurationRepository,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {

	ui = new(testterm.FakeUI)
	cmd := NewTarget(ui, configRepo, orgRepo, spaceRepo)
	ctxt := testcmd.NewContext("target", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
