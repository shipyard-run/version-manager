package gvm

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/google/go-github/github"
	"github.com/hashicorp/go-getter"
	"golang.org/x/xerrors"
)

type Archive int

// Options defines the options for Versions
type Options struct {
	Organization  string
	Repo          string
	GOOS          string // set to os default if blank
	GOARCH        string // set to os value if blank
	AssetNameFunc func(ver, goos, goarch string) string
	ExeNameFunc   func(ver, goos, goarch string) string
	ReleasesPath  string // location to store donwloaded releases
}

// Versions defines the methods for a Go Version Manager implementation
type Versions interface {
	// ListAvailable lists the currently available releases
	// returns a map of version tags with the asset URL
	// Optionally specify a semantic version contstraint to filter results
	// e.g. "~1.2.3", version is greater or equal to 1.2.3 and less than 1.3.0
	ListReleases(constraint string) (map[string]string, error)
	// GetLatestRelease returns the asset for the latest release given the constraint
	GetLatestReleaseURL(constraint string) (tag string, url string, err error)
	// Download and uncompress the release at the given url
	DownloadRelease(tag, url string) (path string, err error)
	// ListInstalledVersions lists versions which have been installed
	ListInstalledVersions(constraint string) (map[string]string, error)
	// GetInstalledVersion returns the version for the latests release given the constraint
	GetInstalledVersion(constraint string) (tag string, path string, err error)
	// SortMapKeys sorts the keys in the map and returns a sorted slice
	// keys must adhere to Semver
	SortMapKeys(map[string]string, bool) []string
	// InRange returns true when the version can be satisfied by the constraint
	// Returns an error if either the constraint or the version are not valid semantic versions
	InRange(version string, constraint string) (bool, error)
}

// New creates a new Versions for the given options
func New(o Options) Versions {
	client := github.NewClient(nil)

	if o.GOARCH == "" {
		o.GOARCH = runtime.GOARCH
	}

	if o.GOOS == "" {
		o.GOOS = runtime.GOOS
	}

	return &VersionsImpl{o, client}
}

// VersionsImpl is the concrete implementation for the Versions interface
type VersionsImpl struct {
	options Options
	client  *github.Client
}

// ListReleases returns a map of assets for releases which match
// the given semantic version and which contain assets matching the value
// returned from AssetNameFunc
// If no version is specified all versions with matching assets are returned
// Release tags which are not valid semantic versions are ignored
func (v *VersionsImpl) ListReleases(constraint string) (map[string]string, error) {
	gr, _, err := v.client.Repositories.ListReleases(context.Background(), v.options.Organization, v.options.Repo, nil)
	if err != nil {
		return nil, xerrors.Errorf("Unable to list Github releases: %w", err)
	}

	tags := map[string]string{}

	for _, g := range gr {
		// does this tag match the provided semver
		if constraint != "" {
			valid, err := v.InRange(*g.TagName, constraint)
			if err != nil {
				return nil, xerrors.Errorf("Invalid sematic version constraint: %w", err)
			}

			// if the tag does not match continue
			if !valid {
				continue
			}
		}

		// check there is an asset with the given filename
		for _, a := range g.Assets {
			tag := strings.TrimLeft(*g.TagName, "v")
			fn := v.options.AssetNameFunc(tag, v.options.GOOS, v.options.GOARCH)
			if strings.ToLower(*a.Name) == strings.ToLower(fn) {
				tags[*g.TagName] = *a.BrowserDownloadURL
				break
			}
		}
	}

	return tags, nil
}

// GetLatestRelease returns the asset which has the latest semantic version matching the constraint
func (v *VersionsImpl) GetLatestReleaseURL(constraint string) (string, string, error) {
	assets, err := v.ListReleases(constraint)
	if err != nil {
		return "", "", err
	}

	keys := v.SortMapKeys(assets, false)

	if len(keys) == 0 {
		return "", "", nil
	}

	tag := keys[len(keys)-1]
	return tag, assets[tag], nil
}

// DownloadRelease and uncompress the given release
func (v *VersionsImpl) DownloadRelease(tag, url string) (filePath string, err error) {
	dir := path.Join(v.options.ReleasesPath, tag)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", xerrors.Errorf("Unable to create temporary folder: %w", err)
	}

	// if the tag is prefixed with a v remove it
	tag = strings.TrimLeft(tag, "v")

	fp := path.Join(dir, v.options.ExeNameFunc(tag, v.options.GOOS, v.options.GOARCH))
	err = getter.GetAny(dir, url)
	if err != nil {
		return "", xerrors.Errorf("Unable to download file: %w", err)
	}

	return fp, nil
}

// ListInstalledVersions lists the versions of the software which are installed int the archive folder
func (v *VersionsImpl) ListInstalledVersions(constraint string) (map[string]string, error) {
	versions := map[string]string{}

	// list folders at the archive loacation matching the semver
	files, err := ioutil.ReadDir(v.options.ReleasesPath)
	if err != nil {
		return nil, xerrors.Errorf("Unable to list releases: %w", err)
	}

	for _, f := range files {
		if constraint != "" {
			valid, err := v.InRange(f.Name(), constraint)
			// if the tag does not match continue
			if err != nil || !valid {
				continue
			}
		}

		tag := strings.TrimLeft(f.Name(), "v")
		versions[f.Name()] = path.Join(v.options.ReleasesPath, f.Name(), v.options.ExeNameFunc(tag, v.options.GOOS, v.options.GOARCH))
	}

	return versions, nil
}

func (v *VersionsImpl) GetInstalledVersion(constraint string) (string, string, error) {
	assets, err := v.ListInstalledVersions(constraint)
	if err != nil {
		return "", "", err
	}

	keys := v.SortMapKeys(assets, false)

	if len(keys) == 0 {
		return "", "", nil
	}

	tag := keys[len(keys)-1]
	return tag, assets[tag], nil
}

func (v *VersionsImpl) SortMapKeys(m map[string]string, decending bool) []string {
	vs := []*semver.Version{}
	for k, _ := range m {
		v, _ := semver.NewVersion(k)
		vs = append(vs, v)
	}

	sort.Sort(semver.Collection(vs))

	versions := []string{}

	// return asccending order
	if !decending {
		for _, v := range vs {
			versions = append(versions, v.Original())
		}
		return versions
	}

	for i := len(vs) - 1; i >= 0; i-- {
		versions = append(versions, vs[i].Original())
	}
	return versions
}

func (v *VersionsImpl) InRange(version string, constraint string) (bool, error) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, xerrors.Errorf("Invalid sematic version constraint: %w", err)
	}

	ver, err := semver.NewVersion(version)
	if err != nil {
		return false, xerrors.Errorf("Invalid sematic version: %w", err)
	}

	return c.Check(ver), nil
}
