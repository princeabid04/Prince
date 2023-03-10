package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectLocator_FindRootModules_WithSingleProject(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{})
	mods := pl.FindRootModules("./testdata/project_locator/single_project")

	require.Len(t, mods, 1)
	assert.Contains(t, mods, RootPath{Path: "./testdata/project_locator/single_project", HasChanges: false})
}

func TestProjectLocator_FindRootModules_WithMultiProject(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{})
	mods := pl.FindRootModules("./testdata/project_locator/multi_project_with_module")

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_with_module/dev"})
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_with_module/prod"})
}

func TestProjectLocator_FindRootModules_WithMultiProject_ExcludeDirs(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		ExcludedSubDirs: []string{"dev"},
	})
	mods := pl.FindRootModules("./testdata/project_locator/multi_project_with_module")

	require.Len(t, mods, 1)
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_with_module/prod"})
}

func TestProjectLocator_FindRootModules_WithMultiProject_WithObjectChanges(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		ChangedObjects: []string{
			"./testdata/project_locator/multi_project_with_module/dev/main.tf",
		},
	})
	mods := pl.FindRootModules("./testdata/project_locator/multi_project_with_module")

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_with_module/dev", HasChanges: true})
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_with_module/prod", HasChanges: false})
}

func TestProjectLocator_FindRootModules_WithMultiProject_WithModuleObjectChanges(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		ChangedObjects: []string{
			"./testdata/project_locator/multi_project_with_module/modules/example/main.tf",
		},
	})
	mods := pl.FindRootModules("./testdata/project_locator/multi_project_with_module")

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_with_module/dev", HasChanges: false})
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_with_module/prod", HasChanges: true})
}

func TestProjectLocator_FindRootModules_WithMultiProjectWithoutProviderBlocks(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		UseAllPaths: true,
	})
	mods := pl.FindRootModules("./testdata/project_locator/multi_project_without_provider_blocks")

	require.Len(t, mods, 3)
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_without_provider_blocks/dev"})
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_without_provider_blocks/prod"})
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_without_provider_blocks/modules/example"})
}

func TestProjectLocator_FindRootModules_WithMultiProjectWithoutProviderBlocks_ExludePaths(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		UseAllPaths: true,
		ExcludedSubDirs: []string{
			"modules/**",
		},
	})
	mods := pl.FindRootModules("./testdata/project_locator/multi_project_without_provider_blocks")

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_without_provider_blocks/dev"})
	assert.Contains(t, mods, RootPath{Path: "testdata/project_locator/multi_project_without_provider_blocks/prod"})
}
