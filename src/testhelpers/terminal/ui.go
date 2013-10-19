package terminal

import (
	"fmt"
	"strings"
	"github.com/codegangsta/cli"
	"cf/configuration"
	"time"
)

type FakeUI struct {
	Outputs []string
	Prompts []string
	PasswordPrompts []string
	Inputs  []string
	FailedWithUsage bool
	ShowConfigurationCalled bool
}

func (ui *FakeUI) Say(message string, args ...interface{}) {
	message =  fmt.Sprintf(message, args...)
	ui.Outputs = append(ui.Outputs,message)
	return
}

func (ui *FakeUI) Warn(message string, args ...interface{}) {
	ui.Say(message,args...)
	return
}

func (ui *FakeUI) Ask(prompt string, args ...interface{}) (answer string) {
	ui.Prompts = append(ui.Prompts, fmt.Sprintf(prompt, args...))
	answer = ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return
}

func (ui *FakeUI) Confirm(prompt string, args ...interface{}) bool {
	response := ui.Ask(prompt, args...)
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	}
	return false
}

func (ui *FakeUI) AskForPassword(prompt string, args ...interface{}) (answer string) {
	ui.PasswordPrompts = append(ui.PasswordPrompts, fmt.Sprintf(prompt, args...))
	answer = ui.Inputs[0]
	ui.Inputs = ui.Inputs[1:]
	return
}

func (ui *FakeUI) Ok() {
	ui.Say("OK")
}

func (ui *FakeUI) Failed(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	ui.Say("FAILED")
	ui.Say(message)
	return
}

func (ui *FakeUI) ConfigFailure(err error) {
	ui.Failed("Error loading config file.\n%s",err.Error())
}

func (ui *FakeUI) FailWithUsage(ctxt *cli.Context, cmdName string) {
	ui.FailedWithUsage = true
	ui.Failed("Incorrect Usage.")
}

func (ui *FakeUI) DumpOutputs() string {
	return "****************************\n" + strings.Join(ui.Outputs, "\n")
}

func (ui *FakeUI) ClearOutputs() {
	ui.Outputs = []string{}
}

func (ui *FakeUI) ShowConfiguration(config *configuration.Configuration) {
	ui.ShowConfigurationCalled = true
}

func (ui FakeUI) LoadingIndication() {
}

func (c FakeUI) Wait(duration time.Duration) {
	time.Sleep(duration)
}

func (ui *FakeUI) showBaseConfig(config *configuration.Configuration) {

}

func (ui *FakeUI) DisplayTable(table [][]string) {

	for _, line := range table {
		output := ""
		for _, value := range line {
			output = output + value + "  "
		}
		ui.Say("%s",output)
	}
}
