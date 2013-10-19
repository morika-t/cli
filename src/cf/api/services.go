package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type PaginatedServiceOfferingResources struct {
	Resources []ServiceOfferingResource
}

type ServiceOfferingResource struct {
	Metadata Metadata
	Entity   ServiceOfferingEntity
}

type ServiceOfferingEntity struct {
	Label            string
	Version          string
	Description      string
	DocumentationUrl string `json:"documentation_url"`
	Provider         string
	ServicePlans     []ServicePlanResource `json:"service_plans"`
}

type ServicePlanResource struct {
	Metadata Metadata
	Entity   ServicePlanEntity
}

type ServicePlanEntity struct {
	Name            string
	ServiceOffering ServiceOfferingResource `json:"service"`
}

type PaginatedServiceInstanceResources struct {
	Resources []ServiceInstanceResource
}

type ServiceInstanceResource struct {
	Metadata Metadata
	Entity   ServiceInstanceEntity
}

type ServiceInstanceEntity struct {
	Name            string
	ServiceBindings []ServiceBindingResource `json:"service_bindings"`
	ServicePlan     ServicePlanResource      `json:"service_plan"`
}

type ServiceBindingResource struct {
	Metadata Metadata
	Entity   ServiceBindingEntity
}

type ServiceBindingEntity struct {
	AppGuid string `json:"app_guid"`
}

type ServiceRepository interface {
	GetServiceOfferings() (offerings []cf.ServiceOffering, apiResponse net.ApiResponse)
	FindInstanceByName(name string) (instance cf.ServiceInstance, apiResponse net.ApiResponse)
	CreateServiceInstance(name string, plan cf.ServicePlan) (identicalAlreadyExists bool, apiResponse net.ApiResponse)
	RenameService(instance cf.ServiceInstance, newName string) (apiResponse net.ApiResponse)
	DeleteService(instance cf.ServiceInstance) (apiResponse net.ApiResponse)
}

type CloudControllerServiceRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerServiceRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerServiceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceRepository) GetServiceOfferings() (offerings []cf.ServiceOffering, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/services?inline-relations-depth=1", repo.config.Target)
	spaceGuid := repo.config.Space.Guid

	if spaceGuid != "" {
		path = fmt.Sprintf("%s/v2/spaces/%s/services?inline-relations-depth=1", repo.config.Target, spaceGuid)
	}

	resources := new(PaginatedServiceOfferingResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range resources.Resources {
		plans := []cf.ServicePlan{}
		for _, p := range r.Entity.ServicePlans {
			plans = append(plans, cf.ServicePlan{Name: p.Entity.Name, Guid: p.Metadata.Guid})
		}
		offerings = append(offerings, cf.ServiceOffering{
			Label:       r.Entity.Label,
			Version:     r.Entity.Version,
			Provider:    r.Entity.Provider,
			Description: r.Entity.Description,
			Guid:        r.Metadata.Guid,
			Plans:       plans,
		})
	}

	return
}

func (repo CloudControllerServiceRepository) FindInstanceByName(name string) (instance cf.ServiceInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=name%s&inline-relations-depth=2", repo.config.Target, repo.config.Space.Guid, "%3A"+name)

	resources := new(PaginatedServiceInstanceResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(resources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Service instance %s not found", name)
		return
	}

	resource := resources.Resources[0]
	serviceOfferingEntity := resource.Entity.ServicePlan.Entity.ServiceOffering.Entity
	instance.Guid = resource.Metadata.Guid
	instance.Name = resource.Entity.Name

	instance.ServicePlan = cf.ServicePlan{
		Name: resource.Entity.ServicePlan.Entity.Name,
		Guid: resource.Entity.ServicePlan.Metadata.Guid,
	}

	instance.ServicePlan.ServiceOffering.Label = serviceOfferingEntity.Label
	instance.ServicePlan.ServiceOffering.DocumentationUrl = serviceOfferingEntity.DocumentationUrl
	instance.ServicePlan.ServiceOffering.Description = serviceOfferingEntity.Description

	instance.ServiceBindings = []cf.ServiceBinding{}

	for _, bindingResource := range resource.Entity.ServiceBindings {
		newBinding := cf.ServiceBinding{
			Url:     bindingResource.Metadata.Url,
			Guid:    bindingResource.Metadata.Guid,
			AppGuid: bindingResource.Entity.AppGuid,
		}
		instance.ServiceBindings = append(instance.ServiceBindings, newBinding)
	}

	return
}

func (repo CloudControllerServiceRepository) CreateServiceInstance(name string, plan cf.ServicePlan) (identicalAlreadyExists bool, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_instances", repo.config.Target)
	data := fmt.Sprintf(
		`{"name":"%s","service_plan_guid":"%s","space_guid":"%s"}`,
		name, plan.Guid, repo.config.Space.Guid,
	)

	apiResponse = repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(data))

	if apiResponse.IsNotSuccessful() && apiResponse.ErrorCode == cf.SERVICE_INSTANCE_NAME_TAKEN {

		serviceInstance, findInstanceApiResponse := repo.FindInstanceByName(name)

		if !findInstanceApiResponse.IsNotSuccessful() &&
			serviceInstance.ServicePlan.Guid == plan.Guid {
			apiResponse = net.ApiResponse{}
			identicalAlreadyExists = true
			return
		}
	}
	return
}

func (repo CloudControllerServiceRepository) RenameService(instance cf.ServiceInstance, newName string) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(`{"name":"%s"}`, newName)
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.Target, instance.Guid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceRepository) DeleteService(instance cf.ServiceInstance) (apiResponse net.ApiResponse) {
	if len(instance.ServiceBindings) > 0 {
		return net.NewApiResponseWithMessage("Cannot delete service instance, apps are still bound to it")
	}
	path := fmt.Sprintf("%s/v2/service_instances/%s", repo.config.Target, instance.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
