package commands

import (
	"cf/api"
	"cf/commands/application"
	"cf/commands/domain"
	"cf/commands/organization"
	"cf/commands/route"
	"cf/commands/service"
	"cf/commands/serviceauthtoken"
	"cf/commands/servicebroker"
	"cf/commands/space"
	"cf/commands/user"
	"cf/configuration"
	"cf/terminal"
	"errors"
)

type Factory interface {
	GetByCmdName(cmdName string) (cmd Command, err error)
}

type ConcreteFactory struct {
	cmdsByName map[string]Command
}

func NewFactory(ui terminal.UI, config *configuration.Configuration, configRepo configuration.ConfigurationRepository, repoLocator api.RepositoryLocator) (factory ConcreteFactory) {
	factory.cmdsByName = make(map[string]Command)

	factory.cmdsByName["api"] = NewApi(ui, config, repoLocator.GetEndpointRepository())
	factory.cmdsByName["app"] = application.NewShowApp(ui, repoLocator.GetAppSummaryRepository())
	factory.cmdsByName["apps"] = application.NewListApps(ui, config, repoLocator.GetAppSummaryRepository())
	factory.cmdsByName["auth"] = NewAuthenticate(ui, configRepo, repoLocator.GetAuthenticationRepository())
	factory.cmdsByName["bind-service"] = service.NewBindService(ui, repoLocator.GetServiceBindingRepository())
	factory.cmdsByName["create-org"] = organization.NewCreateOrg(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["create-service"] = service.NewCreateService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["create-service-auth-token"] = serviceauthtoken.NewCreateServiceAuthToken(ui, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["create-service-broker"] = servicebroker.NewCreateServiceBroker(ui, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["create-space"] = space.NewCreateSpace(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["create-user"] = user.NewCreateUser(ui, repoLocator.GetUserRepository())
	factory.cmdsByName["create-user-provided-service"] = service.NewCreateUserProvidedService(ui, repoLocator.GetUserProvidedServiceInstanceRepository())
	factory.cmdsByName["delete"] = application.NewDeleteApp(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["delete-domain"] = domain.NewDeleteDomain(ui, repoLocator.GetDomainRepository())
	factory.cmdsByName["delete-org"] = organization.NewDeleteOrg(ui, repoLocator.GetOrganizationRepository(), configRepo)
	factory.cmdsByName["delete-route"] = route.NewDeleteRoute(ui, repoLocator.GetRouteRepository())
	factory.cmdsByName["delete-service"] = service.NewDeleteService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["delete-service-auth-token"] = serviceauthtoken.NewDeleteServiceAuthToken(ui, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["delete-service-broker"] = servicebroker.NewDeleteServiceBroker(ui, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["delete-space"] = space.NewDeleteSpace(ui, repoLocator.GetSpaceRepository(), configRepo)
	factory.cmdsByName["delete-user"] = user.NewDeleteUser(ui, repoLocator.GetUserRepository())
	factory.cmdsByName["domains"] = domain.NewListDomains(ui, repoLocator.GetDomainRepository())
	factory.cmdsByName["env"] = application.NewEnv(ui)
	factory.cmdsByName["events"] = application.NewEvents(ui, repoLocator.GetAppEventsRepository())
	factory.cmdsByName["files"] = application.NewFiles(ui, repoLocator.GetAppFilesRepository())
	factory.cmdsByName["login"] = NewLogin(ui, configRepo, repoLocator.GetAuthenticationRepository(), repoLocator.GetEndpointRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["logout"] = NewLogout(ui, configRepo)
	factory.cmdsByName["logs"] = application.NewLogs(ui, repoLocator.GetLogsRepository())
	factory.cmdsByName["marketplace"] = service.NewMarketplaceServices(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["map-domain"] = domain.NewDomainMapper(ui, repoLocator.GetDomainRepository(), true)
	factory.cmdsByName["map-route"] = route.NewRouteMapper(ui, repoLocator.GetRouteRepository(), true)
	factory.cmdsByName["org"] = organization.NewShowOrg(ui)
	factory.cmdsByName["org-users"] = user.NewOrgUsers(ui, repoLocator.GetUserRepository())
	factory.cmdsByName["orgs"] = organization.NewListOrgs(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["passwd"] = NewPassword(ui, repoLocator.GetPasswordRepository(), configRepo)
	factory.cmdsByName["rename"] = application.NewRenameApp(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["rename-org"] = organization.NewRenameOrg(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["rename-service"] = service.NewRenameService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["rename-service-broker"] = servicebroker.NewRenameServiceBroker(ui, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["rename-space"] = space.NewRenameSpace(ui, repoLocator.GetSpaceRepository(), configRepo)
	factory.cmdsByName["reserve-domain"] = domain.NewReserveDomain(ui, repoLocator.GetDomainRepository())
	factory.cmdsByName["reserve-route"] = route.NewReserveRoute(ui, repoLocator.GetRouteRepository())
	factory.cmdsByName["routes"] = route.NewListRoutes(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["service"] = service.NewShowService(ui)
	factory.cmdsByName["service-auth-tokens"] = serviceauthtoken.NewListServiceAuthTokens(ui, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["service-brokers"] = servicebroker.NewListServiceBrokers(ui, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["services"] = service.NewListServices(ui, config, repoLocator.GetServiceSummaryRepository())
	factory.cmdsByName["set-env"] = application.NewSetEnv(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["set-org-role"] = user.NewSetOrgRole(ui, repoLocator.GetUserRepository())
	factory.cmdsByName["set-quota"] = organization.NewSetQuota(ui, repoLocator.GetQuotaRepository())
	factory.cmdsByName["set-space-role"] = user.NewSetSpaceRole(ui, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["share-domain"] = domain.NewShareDomain(ui, repoLocator.GetDomainRepository())
	factory.cmdsByName["space"] = space.NewShowSpace(ui, config)
	factory.cmdsByName["space-users"] = user.NewSpaceUsers(ui, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["spaces"] = space.NewListSpaces(ui, config, repoLocator.GetSpaceRepository())
	factory.cmdsByName["stacks"] = NewStacks(ui, repoLocator.GetStackRepository())
	factory.cmdsByName["target"] = NewTarget(ui, configRepo, repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["unbind-service"] = service.NewUnbindService(ui, repoLocator.GetServiceBindingRepository())
	factory.cmdsByName["unmap-domain"] = domain.NewDomainMapper(ui, repoLocator.GetDomainRepository(), false)
	factory.cmdsByName["unmap-route"] = route.NewRouteMapper(ui, repoLocator.GetRouteRepository(), false)
	factory.cmdsByName["unset-env"] = application.NewUnsetEnv(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["unset-org-role"] = user.NewUnsetOrgRole(ui, repoLocator.GetUserRepository())
	factory.cmdsByName["unset-space-role"] = user.NewUnsetSpaceRole(ui, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["update-service-broker"] = servicebroker.NewUpdateServiceBroker(ui, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["update-service-auth-token"] = serviceauthtoken.NewUpdateServiceAuthToken(ui, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["update-user-provided-service"] = service.NewUpdateUserProvidedService(ui, repoLocator.GetUserProvidedServiceInstanceRepository())
	factory.cmdsByName["users"] = user.NewListUsers(ui, repoLocator.GetUserRepository())

	start := application.NewStart(ui, config, repoLocator.GetApplicationRepository())
	stop := application.NewStop(ui, repoLocator.GetApplicationRepository())
	restart := application.NewRestart(ui, start, stop)

	factory.cmdsByName["start"] = start
	factory.cmdsByName["stop"] = stop
	factory.cmdsByName["restart"] = restart
	factory.cmdsByName["push"] = application.NewPush(ui, start, stop, repoLocator.GetApplicationRepository(), repoLocator.GetDomainRepository(), repoLocator.GetRouteRepository(), repoLocator.GetStackRepository(), repoLocator.GetApplicationBitsRepository())
	factory.cmdsByName["scale"] = application.NewScale(ui, restart, repoLocator.GetApplicationRepository())

	return
}

func (f ConcreteFactory) GetByCmdName(cmdName string) (cmd Command, err error) {
	cmd, found := f.cmdsByName[cmdName]
	if !found {
		err = errors.New("Command not found")
	}
	return
}
