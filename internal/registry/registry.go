package registry

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/git-pkgs/registries"
	_ "github.com/git-pkgs/registries/all" // register all ecosystems
	"github.com/git-pkgs/registries/client"
)

type Client struct {
	reg       registries.Registry
	ecosystem string
}

// Options controls optional registry query behavior.
type Options struct {
	// MinReleaseAge filters out versions published more recently than this
	// duration (i.e. versions released within the last MinReleaseAge). Versions
	// without a known publish time are kept because their age cannot be
	// determined. A zero value disables the filter. Ecosystems that do not
	// publish release timestamps (e.g. homebrew, deno, terraform) are
	// unaffected.
	MinReleaseAge time.Duration
}

func New(ecosystem string) (*Client, error) {
	reg, err := registries.New(ecosystem, "", nil)
	if err != nil {
		return nil, fmt.Errorf("unsupported ecosystem: %s", ecosystem)
	}
	return &Client{reg: reg, ecosystem: ecosystem}, nil
}

func SupportedEcosystems() []string {
	return registries.SupportedEcosystems()
}

func (c *Client) Show(ctx context.Context, name string, opts Options) (*formatter.ShowOutput, error) {
	pkg, err := c.reg.FetchPackage(ctx, name)
	if err != nil {
		return nil, c.mapError(name, err)
	}

	// Resolve the version to report. When a minimum release age is requested,
	// prefer the newest version older than the cutoff instead of the registry's
	// reported latest. Fall back to the registry latest when timestamps are
	// unavailable or fetching versions fails.
	version := pkg.LatestVersion
	if opts.MinReleaseAge > 0 {
		if versions, verr := c.reg.FetchVersions(ctx, name); verr == nil {
			if selected := selectLatest(filterByMinReleaseAge(versions, opts.MinReleaseAge)); selected != nil {
				version = selected.Number
			}
		}
	}

	var author, authorEmail string
	maintainers, merr := c.reg.FetchMaintainers(ctx, name)
	if merr == nil && len(maintainers) > 0 {
		author = maintainers[0].Name
		if author == "" {
			author = maintainers[0].Login
		}
		authorEmail = maintainers[0].Email
	}

	var dependencies []string
	deps, derr := c.reg.FetchDependencies(ctx, name, version)
	if derr == nil {
		for _, dep := range deps {
			if dep.Scope == registries.Runtime {
				dependencies = append(dependencies, dependencyString(dep))
			}
		}
	}

	return &formatter.ShowOutput{
		Name:         pkg.Name,
		Version:      version,
		Description:  pkg.Description,
		Author:       author,
		AuthorEmail:  authorEmail,
		License:      pkg.Licenses,
		HomePage:     pkg.Homepage,
		Dependencies: dependencies,
	}, nil
}

func dependencyString(d registries.Dependency) string {
	scope := ""
	if d.Scope != registries.Runtime {
		scope = string(d.Scope)
	}
	opt := ""
	if d.Optional {
		opt = ",opt"
	}
	extra := ""
	if scope != "" || opt != "" {
		extra = " (" + scope + opt + ")"
	}
	return d.Name + d.Requirements + extra
}

func (c *Client) Latest(ctx context.Context, name string, opts Options) (*formatter.LatestOutput, error) {
	versions, err := c.reg.FetchVersions(ctx, name)
	if err != nil {
		return nil, c.mapError(name, err)
	}

	versions = filterByMinReleaseAge(versions, opts.MinReleaseAge)
	ver := selectLatest(versions)
	if ver == nil {
		return nil, fmt.Errorf("no versions found for package '%s'", name)
	}

	return &formatter.LatestOutput{
		Package: name,
		Version: ver.Number,
	}, nil
}

func (c *Client) Versions(ctx context.Context, name string, opts Options) (*formatter.VersionsOutput, error) {
	versions, err := c.reg.FetchVersions(ctx, name)
	if err != nil {
		return nil, c.mapError(name, err)
	}

	versions = filterByMinReleaseAge(versions, opts.MinReleaseAge)

	items := make([]formatter.VersionItem, 0, len(versions))
	for _, v := range versions {
		if v.Status != "" {
			continue
		}
		items = append(items, formatter.VersionItem{
			Version:    v.Number,
			UploadDate: v.PublishedAt.Format("2006-01-02 15:04:05"),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].UploadDate > items[j].UploadDate
	})

	return &formatter.VersionsOutput{
		Package:  name,
		Versions: items,
	}, nil
}

func (c *Client) BrowseURL(name string) string {
	return c.reg.URLs().Registry(name, "")
}

func (c *Client) mapError(name string, err error) error {
	var httpErr *client.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
		return fmt.Errorf("package '%s' not found", name)
	}
	if errors.Is(err, client.ErrNotFound) {
		return fmt.Errorf("package '%s' not found", name)
	}
	return err
}

// filterByMinReleaseAge removes versions published more recently than minAge
// relative to the current time. Versions without a known publish time are kept.
func filterByMinReleaseAge(versions []registries.Version, minAge time.Duration) []registries.Version {
	return filterByMinReleaseAgeAt(versions, minAge, time.Now())
}

// filterByMinReleaseAgeAt is the time-injectable form used for testing.
func filterByMinReleaseAgeAt(versions []registries.Version, minAge time.Duration, now time.Time) []registries.Version {
	if minAge <= 0 {
		return versions
	}
	cutoff := now.Add(-minAge)
	filtered := make([]registries.Version, 0, len(versions))
	for _, v := range versions {
		if !v.PublishedAt.IsZero() && v.PublishedAt.After(cutoff) {
			continue
		}
		filtered = append(filtered, v)
	}
	return filtered
}

// selectLatest returns the newest non-yanked/retracted/deprecated version,
// preferring versions with a known publish time (sorted newest first). When no
// version has a publish time, list order is preserved. Returns nil when no
// valid version exists. This mirrors [registries.FetchLatestVersion] but
// operates on an already-fetched slice so the min-release-age filter can be
// applied first.
func selectLatest(versions []registries.Version) *registries.Version {
	valid := make([]registries.Version, 0, len(versions))
	for _, v := range versions {
		if v.Status == registries.StatusNone {
			valid = append(valid, v)
		}
	}
	if len(valid) == 0 {
		return nil
	}

	hasTimestamps := false
	for _, v := range valid {
		if !v.PublishedAt.IsZero() {
			hasTimestamps = true
			break
		}
	}
	if hasTimestamps {
		sort.Slice(valid, func(i, j int) bool {
			return valid[i].PublishedAt.After(valid[j].PublishedAt)
		})
	}
	return &valid[0]
}
