package internal

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/distribution/reference"

	gdp "github.com/fgrehm/go-dockerpty"
	gdc "github.com/fsouza/go-dockerclient"
)

type containerConfig struct {
	config     gdc.Config
	hostConfig gdc.HostConfig
}

type containerInfo struct {
	id     string
	cmd    []string
	image  string
	mounts []gdc.HostMount
	merged string
	state  gdc.State
}

func PrepareDebugImage(client *gdc.Client, image string, pullImage bool) (*gdc.Image, error) {
	if client == nil {
		return nil, fmt.Errorf("Empty client was passed to function")
	}

	if pullImage {
		pullDebugImage(client, image)

		img, err := client.InspectImage(image)
		return img, err
	}

	img, err := client.InspectImage(image)

	if err != nil || img == nil {
		pullDebugImage(client, image)
		img, err = client.InspectImage(image)
	}

	return img, err
}

func RunDebugContainer(client *gdc.Client, image *gdc.Image, skipNs bool, id string, cmd []string) error {
	if client == nil || image == nil {
		return fmt.Errorf("Empty client or image was passed to function")
	}
	// inspecting old container
	contInfo, err := getDebugContainerInfo(client, image, id, cmd)
	if err != nil {
		return err
	}

	if !contInfo.state.Running {
		skipNs = true
		fmt.Printf("WARNING: Cannot use container namespaces because its status is %s\n", contInfo.state.Status)
	}

	config := prepareDebugContainerConfig(*contInfo, skipNs)

	// run debug container
	container, err := client.CreateContainer(gdc.CreateContainerOptions{
		Name:       fmt.Sprintf("%s-debug-%d", id, time.Now().Unix()),
		Config:     &config.config,
		HostConfig: &config.hostConfig,
	})
	if err != nil {
		return fmt.Errorf("Cannot create debug container: %s", err)
	}

	defer func() {
		// force remove debug container
		errRC := client.RemoveContainer(gdc.RemoveContainerOptions{
			ID:            container.ID,
			RemoveVolumes: true,
			Force:         true,
		})
		errUM := UnmountOverlay(contInfo.merged)

		if errRC != nil {
			log.Printf("Cannot force remove container '%s': %s", container.Name, err)
		}
		if errUM != nil {
			log.Printf("Cannot unmount merged volume '%s': %s", contInfo.merged, err)
		}
		if errRC != nil || errUM != nil {
			os.Exit(1)
		}
	}()

	// start and attach to container
	err = gdp.Start(client, container, nil)
	if err != nil {
		return fmt.Errorf("Container '%s' has failed to start: %s", container.Name, err)
	}

	return nil
}

func getAuthConfig(registry string) gdc.AuthConfiguration {
	authConfigurations, err := gdc.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return gdc.AuthConfiguration{}
	}

	authConfiguration, ok := authConfigurations.Configs[registry]
	if !ok {
		return gdc.AuthConfiguration{}
	}

	return authConfiguration
}

func prepareMounts(merged string, mounts []gdc.Mount) []gdc.HostMount {
	hostMounts := make([]gdc.HostMount, len(mounts)+1)

	hostMounts[0] = gdc.HostMount{
		Target: "/cont",
		Source: merged,
		Type:   "bind",
	}

	for i, mount := range mounts {
		hostMounts[i+1] = gdc.HostMount{
			ReadOnly: !mount.RW,
			Type:     "bind",
			Source:   mount.Source,
			Target:   fmt.Sprintf("/cont%s", mount.Destination),
		}
	}

	return hostMounts
}

func getDebugContainerInfo(client *gdc.Client, image *gdc.Image, id string, cmd []string) (*containerInfo, error) {
	// inspecting container for debugging
	baseContainer, err := client.InspectContainer(id)
	if err != nil {
		return nil, fmt.Errorf("Cannot inspect container '%s': %s", id, err)
	}
	if baseContainer.GraphDriver == nil || baseContainer.GraphDriver.Name != "overlay2" {
		return nil, fmt.Errorf("This tool can only be used with overlay2 storage driver containers")
	}

	merged, err := MountOverlay(
		baseContainer.GraphDriver.Data["LowerDir"],
		baseContainer.GraphDriver.Data["UpperDir"],
		baseContainer.GraphDriver.Data["WorkDir"],
	)

	if err != nil {
		return nil, fmt.Errorf("Cannot prepare merged dir: %s", err)
	}

	hostMounts := prepareMounts(merged, baseContainer.Mounts)

	return &containerInfo{
		id:     id,
		image:  image.ID,
		cmd:    cmd,
		mounts: hostMounts,
		merged: merged,
		state:  baseContainer.State,
	}, nil
}

func prepareDebugContainerConfig(contInfo containerInfo, skipNs bool) containerConfig {
	mode := fmt.Sprintf("container:%s", contInfo.id)

	if skipNs {
		mode = ""
	}

	return containerConfig{
		config: gdc.Config{
			Image:        contInfo.image,
			Cmd:          contInfo.cmd,
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
		hostConfig: gdc.HostConfig{
			CapAdd:      []string{"sys_ptrace", "sys_admin"},
			PidMode:     mode,
			NetworkMode: mode,
			IpcMode:     mode,
			Mounts:      contInfo.mounts,
		},
	}
}

func pullDebugImage(client *gdc.Client, image string) error {
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("Cannot parse image name '%s': %s", image, err)
	}

	err = client.PullImage(gdc.PullImageOptions{
		Repository:   reference.TagNameOnly(named).String(),
		OutputStream: os.Stderr,
	}, getAuthConfig(reference.Domain(named)))

	if err != nil {
		return fmt.Errorf("Cannot pull image '%s': %s", image, err)
	}

	return nil
}
