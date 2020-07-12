package gvm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (string, *VersionsImpl) {
	dlPath, _ := ioutil.TempDir("", "")

	// clean up
	t.Cleanup(func() {
		os.RemoveAll(dlPath)
	})

	o := Options{
		Organization: "nicholasjackson",
		Repo:         "fake-service",
		ReleasesPath: dlPath,
		GOOS:         "linux",
		GOARCH:       "x64",
	}

	nf := func(ver, goos, goarch string) string {
		baseName := "fake-service"

		switch goos {
		case "darwin":
			return fmt.Sprintf("%s-osx", baseName)
		case "linux":
			return fmt.Sprintf("%s-linux", baseName)
		case "windows":
			return fmt.Sprintf("%s.exe", baseName)
		}

		return ""
	}

	o.AssetNameFunc = nf
	o.ExeNameFunc = nf

	v := New(o)

	return dlPath, v.(*VersionsImpl)
}

func TestListReleasesGetsFromGitHub(t *testing.T) {
	_, v := setup(t)

	r, err := v.ListReleases("")
	assert.NoError(t, err)

	assert.Contains(t, r, "v0.14.1")
}

func TestDoesNotListReleasesFromGitHubWhenInvalidConstraint(t *testing.T) {
	_, v := setup(t)

	_, err := v.ListReleases("abd")
	assert.Error(t, err)
}

func TestDoesNotListReleasesFromGitHubWhenNoReleasesForConstraint(t *testing.T) {
	_, v := setup(t)

	r, err := v.ListReleases("~1.0.0")
	assert.NoError(t, err)

	assert.Len(t, r, 0)
}

func TestDoesNotListReleasesFromGitHubUnknownArch(t *testing.T) {
	_, v := setup(t)
	v.options.GOOS = "fake"

	r, err := v.ListReleases("")
	assert.NoError(t, err)

	assert.NotContains(t, r, "v0.14.1")
}

func TestGetLatestReleasesGetsFromGitHub(t *testing.T) {
	_, v := setup(t)

	tag, url, err := v.GetLatestReleaseURL("~v0.12.0")
	assert.NoError(t, err)

	assert.Contains(t, url, "v0.12.2")
	assert.Equal(t, tag, "v0.12.2")
}

func TestDownloadsLatestReleasesFromGitHub(t *testing.T) {
	_, v := setup(t)

	tag, url, err := v.GetLatestReleaseURL("~v0.12.0")
	assert.NoError(t, err)

	dl, err := v.DownloadRelease(tag, url)
	assert.NoError(t, err)

	assert.FileExists(t, dl)
}

func TestListInstalledReturnsVersions(t *testing.T) {
	tmp, v := setup(t)

	fn := path.Join(tmp, "v0.14.1", "fake-service-linux")
	os.MkdirAll(path.Join(tmp, "v0.14.1"), os.ModePerm)
	os.Create(fn)

	r, err := v.ListInstalledVersions("")
	assert.NoError(t, err)

	assert.Equal(t, fn, r["v0.14.1"])
}

func TestListInstalledReturnsNothingWithInvalidVersionSemver(t *testing.T) {
	tmp, v := setup(t)

	os.MkdirAll(path.Join(tmp, "v0.14.1"), os.ModePerm)
	os.Create(path.Join(tmp, "v0.14.1", "fake-service-linux"))

	r, err := v.ListInstalledVersions("v0.15.1")
	assert.NoError(t, err)

	assert.NotContains(t, r, "v0.14.1")
}

func TestListInstalledReturnsLatestVersion(t *testing.T) {
	tmp, v := setup(t)

	os.MkdirAll(path.Join(tmp, "v0.14.1"), os.ModePerm)
	os.Create(path.Join(tmp, "v0.14.1", "fake-service-linux"))
	os.MkdirAll(path.Join(tmp, "v0.14.2"), os.ModePerm)
	os.Create(path.Join(tmp, "v0.14.2", "fake-service-linux"))

	r, err := v.ListInstalledVersions("~v0.14.0")
	assert.NoError(t, err)

	assert.Contains(t, r, "v0.14.2")
}
