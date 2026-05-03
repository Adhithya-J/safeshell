package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerAPI interface {
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
}

type Runner struct {
	cli DockerAPI
	img string
}

func NewRunner(image string) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Runner{cli: cli, img: image}, nil
}

func (r *Runner) SetClient(cli DockerAPI) {
	r.cli = cli
}

func (r *Runner) RunScript(ctx context.Context, script string) (string, error) {
	// 1. Pull Image
	reader, err := r.cli.ImagePull(ctx, r.img, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader) // Show pull progress (Optional, might be noisy)

	// 2. Create Container
	resp, err := r.cli.ContainerCreate(ctx, &container.Config{
		Image: r.img,
		Cmd:   []string{"sh", "-c", script},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// 5. Cleanup
	defer func() {
		r.cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	}()

	// 3. Start Container
	if err := r.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// 4. Wait & Logs
	statusCh, errCh := r.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("error waiting for container: %w", err)
		}
	case <-statusCh:
	}

	out, err := r.cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	defer out.Close()

	var outputBuffer bytes.Buffer
	// MultiWriter to both show output in console and capture it
	mwStdout := io.MultiWriter(os.Stdout, &outputBuffer)
	mwStderr := io.MultiWriter(os.Stderr, &outputBuffer)

	_, err = stdcopy.StdCopy(mwStdout, mwStderr, out)
	if err != nil {
		return "", fmt.Errorf("failed to copy logs: %w", err)
	}
	return outputBuffer.String(), nil // fix this and see how output can be printed
}
