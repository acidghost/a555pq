package registry

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/git-pkgs/registries"
	_ "github.com/git-pkgs/registries/all" // register all ecosystems
	"github.com/git-pkgs/registries/client"
)

type Client struct {
	reg       registries.Registry
	ecosystem string
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

func (c *Client) Show(ctx context.Context, name string) (*formatter.ShowOutput, error) {
	pkg, err := c.reg.FetchPackage(ctx, name)
	if err != nil {
		return nil, c.mapError(name, err)
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
	deps, derr := c.reg.FetchDependencies(ctx, name, pkg.LatestVersion)
	if derr == nil {
		for _, dep := range deps {
			if dep.Scope == registries.Runtime {
				dependencies = append(dependencies, dependencyString(dep))
			}
		}
	}

	return &formatter.ShowOutput{
		Name:         pkg.Name,
		Version:      pkg.LatestVersion,
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

func (c *Client) Latest(ctx context.Context, name string) (*formatter.LatestOutput, error) {
	ver, err := registries.FetchLatestVersion(ctx, c.reg, name)
	if err != nil {
		return nil, c.mapError(name, err)
	}

	return &formatter.LatestOutput{
		Package: name,
		Version: ver.Number,
	}, nil
}

func (c *Client) Versions(ctx context.Context, name string) (*formatter.VersionsOutput, error) {
	versions, err := c.reg.FetchVersions(ctx, name)
	if err != nil {
		return nil, c.mapError(name, err)
	}

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
