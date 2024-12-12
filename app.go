package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
	"golang.org/x/tools/go/packages"
)

// Deps is a set of dependencies for testing
type Deps struct {
	stdout io.Writer
	loader PackagesLoader
}

func NewDeps(stdout io.Writer, loader PackagesLoader) *Deps {
	return &Deps{
		stdout: stdout,
		loader: loader,
	}
}

// PackagesLoader is an interface to get packages.Load testable
type PackagesLoader interface {
	PackagesLoad(cfg *packages.Config, patterns ...string) ([]*packages.Package, error)
}

// X is a wrapper for x/tools/go/packages.Load
type X struct{}

func (x *X) PackagesLoad(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
	return packages.Load(cfg, patterns...)
}

type WhyGoOverResults map[string][]string

func (r WhyGoOverResults) String() string {
	iter := slices.SortedFunc(maps.Keys(r), semver.Compare)

	var sb strings.Builder
	for _, k := range iter {
		v := r[k]
		sb.WriteString(fmt.Sprintf("# Go %s\n", k))
		for _, vv := range v {
			sb.WriteString(fmt.Sprintf("%s\n", vv))
		}
	}
	return sb.String()
}

// WhyGoOver is the main function of this application
func (d *Deps) WhyGoOver(ctx context.Context, currentModulePath, thresholdGoVersion string) (WhyGoOverResults, error) {
	threshold, ok := ensureSemver(thresholdGoVersion)
	if !ok {
		return nil, fmt.Errorf("release version %s is not a valid semver", thresholdGoVersion)
	}

	cfg := &packages.Config{
		Mode: packages.NeedModule | packages.NeedName,
	}

	pkgs, err := d.loader.PackagesLoad(cfg, "all")
	if err != nil {
		return nil, fmt.Errorf("error PackagesLoad: %w", err)
	}
	slog.DebugContext(ctx, "PackagesLoad", slog.Int("len", len(pkgs)))
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no dependencies found: maybe you are not in a go module project")
	}

	results := make(map[string][]string)
	for _, pkg := range pkgs {
		if pkg.Module == nil || pkg.Module.Path == currentModulePath || pkg.Module.GoVersion == "" {
			slog.DebugContext(ctx, "skip", slog.String("pkg", pkg.PkgPath))
			continue
		}

		moduleWithVersion := fmt.Sprintf("%s@%s", pkg.Module.Path, pkg.Module.Version)

		slog.DebugContext(ctx, fmt.Sprintf("Go version %s required", pkg.Module.GoVersion), slog.String("module", moduleWithVersion))
		requiredGoVersion, ok := ensureSemver(pkg.Module.GoVersion)
		if !ok {
			slog.WarnContext(ctx, fmt.Sprintf("Go version `%s` is not a valid semver", pkg.Module.GoVersion), slog.String("module", moduleWithVersion))
			continue
		}

		if semver.Compare(requiredGoVersion, threshold) <= 0 {
			slog.DebugContext(ctx, "not over threshold", slog.String("module", moduleWithVersion), slog.String("required_go", requiredGoVersion))
			continue
		}

		slog.DebugContext(ctx, "over threshold", slog.String("module", moduleWithVersion), slog.String("required_go", requiredGoVersion))
		existingModules := results[requiredGoVersion]

		if !slices.Contains(existingModules, moduleWithVersion) {
			results[requiredGoVersion] = append(existingModules, moduleWithVersion)
		}
	}

	return results, nil
}

func ensureSemver(arg string) (string, bool) {
	result := arg
	if !strings.HasPrefix(arg, "v") {
		result = "v" + result
	}
	if !semver.IsValid(result) {
		return "", false
	}
	return result, true
}
