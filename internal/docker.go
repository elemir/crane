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
	id string
	cmd []string
	image string
	mounts []gdc.HostMount
}

func PrepareDebugImage(client *gdc.Client, image string, pullImage bool) (*gdc.Image, error) {
	if client == nil {
        return nil, fmt.Errorf("Empty client was passed to fun")
	}

	if (pullImage) {
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

func RunDebugContainer(client *gdc.Client, image *gdc.Image, id string, cmd []string) error {
	if client == nil || image == nil {
        return fmt.Errorf("Empty client or image was passed to fun")
	}
	// inspecting old container
	contInfo, err := getDebugContainerInfo(client, image, id, cmd)
	if err != nil {
		return err
	}

	config := prepareDebugContainerConfig(*contInfo)

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
		err = client.RemoveContainer(gdc.RemoveContainerOptions{
			ID:            container.ID,
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			log.Fatalf("Cannot force remove container '%s': %s", container.Name, err)
		}
	}()

	// start and attach to container
	err = gdp.Start(client, container, nil)
	if err != nil {
		return fmt.Errorf("Starting container '%s' was failed: %s", container.Name, err)
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

func prepareMounts(vol string, mounts []gdc.Mount) []gdc.HostMount {
	hostMounts := make([]gdc.HostMount, len(mounts)+1)

	hostMounts[0] = gdc.HostMount{
       Target: "/cont",
       Source: vol,
       Type:   "bind",
	}

	for i, mount := range mounts {
		hostMounts[i+1] = gdc.HostMount{
            ReadOnly: !mount.RW,
			Type: "bind",
			Source: mount.Source,
			Target: fmt.Sprintf("/cont%s", mount.Destination),
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
		return nil, fmt.Errorf("This tool is useful only with overlay2 storage driver containers")
	}

    hostMounts := prepareMounts(baseContainer.GraphDriver.Data["MergedDir"], baseContainer.Mounts)

	return &containerInfo{
		id: id,
		image: image.ID,
		cmd: cmd,
		mounts: hostMounts,
	}, nil
}

func prepareDebugContainerConfig(contInfo containerInfo) containerConfig {
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
			PidMode:     fmt.Sprintf("container:%s", contInfo.id),
			NetworkMode: fmt.Sprintf("container:%s", contInfo.id),
			IpcMode:     fmt.Sprintf("container:%s", contInfo.id),
			Mounts:      contInfo.mounts,
		},
	}
}

func pullDebugImage(client *gdc.Client, image string) error {
    named, err := reference.ParseNormalizedNamed(image)
    if err != nil {
        return fmt.Errorf("Cannot parse image name '%s': %s", image, err)
	}

	err = client.PullImage(gdc.PullImageOptions {
		Repository: reference.TagNameOnly(named).String(),
		OutputStream: os.Stderr,
	}, getAuthConfig(reference.Domain(named)))

	if err != nil {
		return fmt.Errorf("Cannot pull image '%s': %s", image, err)
	}

	return nil
}


