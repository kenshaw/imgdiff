// Command imgdiff is a command-line image diff tool.
package main

import (
	"context"
	"fmt"
	"image"
	"os"

	"github.com/kenshaw/colors"
	"github.com/kenshaw/rasterm"
	"github.com/orisano/pixelmatch"
	"github.com/spf13/cobra"
	"github.com/sunshineplan/imgconv"
)

var (
	name    = "imgdiff"
	version = "0.0.0-dev"
)

func main() {
	if err := run(context.Background(), name, version, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, appName, appVersion string, cliargs []string) error {
	diff := colors.Red.Color()
	c := &cobra.Command{
		Use:     appName + " [flags] <image1> <image2> [image3, ..., imageN]",
		Short:   appName + ", a image diff tool",
		Version: appVersion,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !rasterm.Available() {
				return rasterm.ErrTermGraphicsNotAvailable
			}
			// open first image
			a, err := imgconv.Open(args[0])
			if err != nil {
				return fmt.Errorf("unable to open %s: %v", args[0], err)
			}
			// successively compare all passed images
			for _, s := range args[1:] {
				b, err := imgconv.Open(s)
				if err != nil {
					return fmt.Errorf("unable to open %s: %v", s, err)
				}
				var img image.Image
				if _, err := pixelmatch.MatchPixel(
					a, b,
					pixelmatch.DiffColor(diff),
					pixelmatch.EnableDiffMask,
					pixelmatch.WriteTo(&img),
				); err != nil {
					return fmt.Errorf("unable to compare %s with %s: %v", args[0], s, err)
				}
				if img != nil {
					if err := rasterm.Encode(os.Stdout, img); err != nil {
						return fmt.Errorf("unable to write comparison of %s with %s: %v", args[0], s, err)
					}
				}
			}
			return nil
		},
	}
	c.Flags().Var(diff.Pflag(), "diff-color", "diff color")
	c.SetVersionTemplate("{{ .Name }} {{ .Version }}\n")
	c.InitDefaultHelpCmd()
	c.SetArgs(cliargs[1:])
	c.SilenceErrors, c.SilenceUsage = true, false
	return c.ExecuteContext(ctx)
}
