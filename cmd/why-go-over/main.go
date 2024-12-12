package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/alecthomas/kong"
	app "github.com/ebi-yade/why-go-over"
	"golang.org/x/mod/modfile"
)

var CLI = struct {
	ReleaseVersion string `arg:"" name:"release-version" help:"A release version of Go but RCs"`
	Debug          bool   `name:"debug" env:"WGO_DEBUG"`

	V VersionFlag `short:"v" help:"Show the version of why-go-over"`
}{}

func main() {
	if err := _main(); err != nil {
		slog.Error(fmt.Sprintf("%v", err))
	}
}

func _main() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	kong.Parse(&CLI)

	if CLI.Debug {
		level := slog.LevelDebug
		slog.SetLogLoggerLevel(level)
		slog.DebugContext(ctx, "SetLogLoggerLevel", slog.String("level", level.String()))
	}

	currentModulePath, err := getCurrentModulePath()
	if err != nil {
		return err
	}

	deps := app.NewDeps(&app.X{})
	res, err := deps.WhyGoOver(ctx, currentModulePath, CLI.ReleaseVersion)
	if err != nil {
		return err
	}

	fmt.Print(res)

	return nil
}

func getCurrentModulePath() (string, error) {
	modFile, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("error reading go.mod: %w", err)
	}
	parsed, err := modfile.Parse("go.mod", modFile, nil)
	if err != nil {
		return "", fmt.Errorf("error parsing go.mod: %w", err)
	}

	return parsed.Module.Mod.Path, nil
}

// =================== Printing the tool version ===================

var Version = "(dev)"

type VersionFlag bool

func (v VersionFlag) BeforeReset(app *kong.Kong) error {
	fmt.Fprintf(app.Stdout, "why-go-over %s\n", Version)
	app.Exit(0)
	return nil
}
