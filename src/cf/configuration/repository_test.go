package configuration

import (
	"cf"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadingWithNoConfigFile(t *testing.T) {
	repo := NewConfigurationDiskRepository()
	config := repo.loadDefaultConfig(t)
	defer repo.restoreConfig(t)

	assert.Equal(t, config.Target, "https://api.run.pivotal.io")
	assert.Equal(t, config.ApiVersion, "2")
	assert.Equal(t, config.AuthorizationEndpoint, "https://login.run.pivotal.io")
	assert.Equal(t, config.AccessToken, "")
}

func TestSavingAndLoading(t *testing.T) {
	repo := NewConfigurationDiskRepository()
	configToSave := repo.loadDefaultConfig(t)
	defer repo.restoreConfig(t)

	configToSave.ApiVersion = "3.1.0"
	configToSave.Target = "https://api.target.example.com"
	configToSave.AuthorizationEndpoint = "https://login.target.example.com"
	configToSave.AccessToken = "bearer my_access_token"

	repo.Save()

	singleton = nil
	savedConfig, err := repo.Get()
	assert.NoError(t, err)
	assert.Equal(t, savedConfig, configToSave)
}

func TestSetOrganization(t *testing.T) {
	repo := NewConfigurationDiskRepository()
	config := repo.loadDefaultConfig(t)
	defer repo.restoreConfig(t)

	config.Space = cf.Space{Guid: "my-space-guid"}
	config.Organization = cf.Organization{}

	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	err := repo.SetOrganization(org)
	assert.NoError(t, err)

	repo.Save()

	savedConfig, err := repo.Get()
	assert.NoError(t, err)
	assert.Equal(t, savedConfig.Organization, org)
	assert.Equal(t, savedConfig.Space, cf.Space{})
}

func TestSetSpace(t *testing.T) {
	repo := NewConfigurationDiskRepository()
	repo.loadDefaultConfig(t)
	defer repo.restoreConfig(t)

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	err := repo.SetSpace(space)
	assert.NoError(t, err)

	repo.Save()

	savedConfig, err := repo.Get()
	assert.NoError(t, err)
	assert.Equal(t, savedConfig.Space, space)
}

func (repo ConfigurationDiskRepository) loadDefaultConfig(t *testing.T) (config *Configuration) {
	file, err := ConfigFile()
	assert.NoError(t, err)

	_, err = os.Stat(file)
	if !os.IsNotExist(err) {
		err = os.Rename(file, file+"test-backup")
		assert.NoError(t, err)
	}

	config, err = repo.Get()
	assert.NoError(t, err)

	return
}

func (repo ConfigurationDiskRepository) restoreConfig(t *testing.T) {
	file, err := ConfigFile()
	assert.NoError(t, err)

	err = os.Remove(file)
	assert.NoError(t, err)

	_, err = os.Stat(file + "test-backup")
	if !os.IsNotExist(err) {
		err = os.Rename(file+"test-backup", file)
		assert.NoError(t, err)
	}

	return
}
