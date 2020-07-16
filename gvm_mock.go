package gvm

import (
	"github.com/stretchr/testify/mock"
)

type MockVersions struct {
	mock.Mock
}

func (m *MockVersions) ListReleases(constraint string) (map[string]string, error) {
	args := m.Called(constraint)

	if ma, ok := args.Get(0).(map[string]string); ok {
		return ma, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockVersions) GetLatestReleaseURL(constraint string) (tag string, url string, err error) {
	args := m.Called(constraint)

	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockVersions) DownloadRelease(tag, url string) (path string, err error) {
	args := m.Called(tag, url)

	return args.String(0), args.Error(1)
}

func (m *MockVersions) ListInstalledVersions(constraint string) (map[string]string, error) {
	args := m.Called(constraint)

	if ma, ok := args.Get(0).(map[string]string); ok {
		return ma, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockVersions) GetInstalledVersion(constraint string) (tag string, path string, err error) {
	args := m.Called(constraint)

	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockVersions) SortMapKeys(ma map[string]string, descending bool) []string {
	args := m.Called(ma, descending)

	if rm, ok := args.Get(0).([]string); ok {
		return rm
	}

	return nil
}

func (m *MockVersions) InRange(version string, constraint string) (bool, error) {
	args := m.Called(version, constraint)

	return args.Bool(0), args.Error(1)
}
