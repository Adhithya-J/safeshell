package container

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type mockDockerAPI struct {
	ImagePullFunc       func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreateFunc func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStartFunc  func(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerWaitFunc   func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerLogsFunc   func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerRemoveFunc func(ctx context.Context, containerID string, options container.RemoveOptions) error
}

func (m *mockDockerAPI) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	return m.ImagePullFunc(ctx, ref, options)
}

func (m *mockDockerAPI) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	return m.ContainerCreateFunc(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func (m *mockDockerAPI) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return m.ContainerStartFunc(ctx, containerID, options)
}

func (m *mockDockerAPI) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	return m.ContainerWaitFunc(ctx, containerID, condition)
}

func (m *mockDockerAPI) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	return m.ContainerLogsFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return m.ContainerRemoveFunc(ctx, containerID, options)
}

func TestRunScript_Success(t *testing.T) {
	runner := &Runner{img: "alpine:latest"}

	calls := []string{}
	mock := &mockDockerAPI{
		ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
			calls = append(calls, "Pull")
			return io.NopCloser(bytes.NewBufferString("pulling")), nil
		},
		ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
			calls = append(calls, "Create")
			return container.CreateResponse{ID: "test-id"}, nil
		},
		ContainerStartFunc: func(ctx context.Context, containerID string, options container.StartOptions) error {
			calls = append(calls, "Start")
			return nil
		},
		ContainerWaitFunc: func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
			calls = append(calls, "Wait")
			statusCh := make(chan container.WaitResponse, 1)
			statusCh <- container.WaitResponse{StatusCode: 0}
			return statusCh, nil
		},
		ContainerLogsFunc: func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
			calls = append(calls, "Logs")
			// Docker logs have a specific header (8 bytes): [STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4]
			// 1 for stdout, 2 for stderr
			content := "output"
			header := []byte{1, 0, 0, 0, 0, 0, 0, byte(len(content))}
			return io.NopCloser(bytes.NewBuffer(append(header, []byte(content)...))), nil
		},
		ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			calls = append(calls, "Remove")
			return nil
		},
	}
	runner.SetClient(mock)

	out, err := runner.RunScript(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if out != "output" {
		t.Errorf("Expected output 'output', got '%s'", out)
	}

	expectedCalls := []string{"Pull", "Create", "Start", "Wait", "Logs", "Remove"}

	if len(calls) != len(expectedCalls) {
		t.Errorf("Expected %d calls, got %d: %v", len(expectedCalls), len(calls), calls)
	}

	for i, expected := range expectedCalls {
		if i < len(calls) && calls[i] != expected {
			t.Errorf("Expected call %d to be %s, got %s", i, expected, calls[i])
		}
	}
}
