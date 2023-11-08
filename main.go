// Command imgdiff is a command-line image diff tool.
package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"io"
	"os"

	"github.com/BourgeoisBear/rasterm"
	"github.com/orisano/pixelmatch"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/sunshineplan/imgconv"
	"gopkg.in/go-playground/colors.v1"
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
	var diffColor colors.Color = colors.FromStdColor(color.RGBA{R: 255})
	c := &cobra.Command{
		Use:     appName + " [flags] <image1> <image2> [image3, ..., imageN]",
		Short:   appName + ", a image diff tool",
		Version: appVersion,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, _, ok := renderer()
			if !ok {
				return errors.New("terminal does not support graphics")
			}
			dc := diffColor.ToRGB()
			clr := color.RGBA{R: dc.R, G: dc.G, B: dc.B}
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
					pixelmatch.DiffColor(clr),
					pixelmatch.EnableDiffMask,
					pixelmatch.WriteTo(&img),
				); err != nil {
					return fmt.Errorf("unable to compare %s with %s: %v", args[0], s, err)
				}
				if err := r(os.Stdout, palettize(img)); err != nil {
					return fmt.Errorf("unable to write comparison of %s with %s: %v", args[0], s, err)
				}
			}
			return nil
		},
	}
	c.Flags().Var(NewColor(&diffColor), "diff-color", "diff color")
	c.SetVersionTemplate("{{ .Name }} {{ .Version }}\n")
	c.InitDefaultHelpCmd()
	c.SetArgs(cliargs[1:])
	c.SilenceErrors, c.SilenceUsage = true, false
	return c.ExecuteContext(ctx)
}

func renderer() (func(io.Writer, image.Image) error, string, bool) {
	var s rasterm.Settings
	switch sixel, _ := rasterm.IsSixelCapable(); {
	case rasterm.IsTmuxScreen():
		return nil, "", false
	case rasterm.IsTermKitty():
		return s.KittyWriteImage, "kitty", true
	case rasterm.IsTermItermWez():
		return s.ItermWriteImage, "iterm", true
	case sixel:
		return func(w io.Writer, img image.Image) error {
			pi, ok := img.(*image.Paletted)
			if !ok {
				return errors.New("invalid image")
			}
			return s.SixelWriteImage(w, pi)
		}, "sixel", true
	}
	return nil, "", false
}

func palettize(src image.Image) *image.Paletted {
	if pi, ok := src.(*image.Paletted); ok {
		return pi
	}
	b := src.Bounds()
	img := image.NewPaletted(b, palette.Plan9)
	draw.FloydSteinberg.Draw(img, b, src, image.Point{})
	return img
}

type Color struct {
	c *colors.Color
}

func NewColor(c *colors.Color) pflag.Value {
	return Color{
		c: c,
	}
}

func (c Color) String() string {
	return (*c.c).String()
}

func (c Color) Set(s string) error {
	var err error
	*c.c, err = colors.Parse(s)
	return err
}

func (c Color) Type() string {
	return "color"
}
