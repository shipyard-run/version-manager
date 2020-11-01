# Version Manager 

![Go](https://github.com/shipyard-run/version-manager/workflows/Go/badge.svg)

Version manager allows you to manage and download the latest software versions for your applications from GitHub Releases.

## Example use

### Creating an instance of Version Manager

First set the archive location

```go
dlPath, _ := ioutil.TempDir("", "")
```

Then set the options for the GitHub organization and the repository, the architecture and operating system are 
determined at runtime, you can override these values.

```go
	o := Options{
		Organization: "nicholasjackson",
		Repo:         "fake-service",
		ReleasesPath: dlPath,
		//GOOS:         "linux",
		//GOARCH:       "x64",
	}
```

To determine the name of the asset you define a function, tools like GoReleaser will use formularic naming
for assets added to a release. The asset name may be an archive, in this instance Version Manager will automatically unpack the archive.

```go
	nf := func(goos, goarch string) string {
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
```

In the instance that the release asset is an archive you can define a second function to determine the name of the executable within the archive.
In the instance that the asset is a simple binary you can set this to the same function as previously used.

```go
	o.ExeNameFunc = nf
```

You can now create a new instance of Version Manager

```go
  v := New(o)
```

### Listing Releases on GitHub

To list available releases you can use the `ListReleases` method, this takes a single parameter of a Semantic Version constraint.
For example, to list all the releases `>= 1.2.3, < 2.0.0`, you would specify a constraint of `^1.2.3`. Version manager returns you a map
of release name and asset download URL.

```
r, err := v.ListReleases("^1.2.3")
for version, url := range r {
  fmt.Println("Version:", version, "URL:", url)
}
```

### Downloading the latest version matching the constraint

To download the latest version which matches the given constraint.

```go
tag, url, err := v.GetLatestReleaseURL("~v0.12.0")
assert.NoError(t, err)

dl, err := v.DownloadRelease(tag, url)
assert.NoError(t, err)
```
