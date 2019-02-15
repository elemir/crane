package main

import (
	"fmt"
	"log"

	"github.com/elemir/rainforest"
	"github.com/spf13/cobra"

	"github.com/elemir/crane/internal"

	gdc "github.com/fsouza/go-dockerclient"
)

var DefaultRepo = "elemir"
var DefaultImage = "crane"
var DefaultTag = "latest"
var skipNs bool
var pull bool
var image string

var rootCmd = &cobra.Command{
	Use:   "crane CONTAINER [COMMAND] [ARG...]",
	Short: "crane is a small utility for debugging docker containers",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := gdc.NewClientFromEnv()
		if err != nil {
			log.Fatalf("Cannot connect to docker: %s", err)
		}
		img, err := internal.PrepareDebugImage(client, image, pull)
		if err != nil {
			log.Fatalf("Cannot prepare debug image: %s", err)
		}
		err = internal.RunDebugContainer(client, img, skipNs, args[0], args[1:])
		if err != nil {
			log.Fatalf("Problems with debug container: %s", err)
		}
	},
}

func main() {
	rootCmd.PersistentFlags().StringVar(&image, "image", fmt.Sprintf("%s/%s:%s", DefaultRepo, DefaultImage, DefaultTag), "Image with debugging tools")
	rootCmd.PersistentFlags().BoolVar(&skipNs, "skip-ns", false, "Skip namespace separation")
	rootCmd.PersistentFlags().BoolVar(&pull, "pull", false, "Always attempt to pull a newer version of the image")
	rootCmd.Flags().SetInterspersed(false)
	rainforest.BindPFlag("crane_image", rootCmd.PersistentFlags().Lookup("image"))
	rainforest.Load()
	rootCmd.Execute()
}
