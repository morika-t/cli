package space

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListSpaces struct {
	ui        terminal.UI
	config    *configuration.Configuration
	spaceRepo api.SpaceRepository
}

func NewListSpaces(ui terminal.UI, config *configuration.Configuration, spaceRepo api.SpaceRepository) (cmd ListSpaces) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd ListSpaces) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd ListSpaces) Run(c *cli.Context) {
	cmd.ui.Say("Getting spaces in org %s as %s...",
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Username()))

	spaces, apiResponse := cmd.spaceRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	for _, space := range spaces {
		cmd.ui.Say(space.Name)
	}
}
