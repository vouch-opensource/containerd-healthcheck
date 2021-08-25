package containerd

import (
	"context"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/sirupsen/logrus"
)

type Containerd struct {
	Logger  *logrus.Logger
	Client  *containerd.Client
	Context context.Context
}

func NewClient(logger *logrus.Logger, socketPath string, namespace string) (*Containerd, error) {
	client, err := containerd.New(socketPath)
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		logger.Error("Unable to connect to the containerd socket:", err)
		return nil, err
	}
	return &Containerd{
		Client:  client,
		Context: ctx,
		Logger:  logger,
	}, nil
}

func (c *Containerd) RestartTask(containerName string) error {
	ctx := c.Context
	container, err := c.Client.LoadContainer(ctx, containerName)
	if err != nil {
		c.Logger.Error("Unable to connect to the containerd socket")
		return err
	}

	if err := c.stopTask(container); err != nil {
		return err
	}

	if err := c.startTask(container); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	return nil

}

func (c *Containerd) startTask(container containerd.Container) error {
	ctx := c.Context
	task, err := container.NewTask(ctx, cio.NullIO)
	if err != nil {
		return err
	}
	if err := task.Start(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Containerd) stopTask(container containerd.Container) error {
	ctx := c.Context
	task, err := container.Task(ctx, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "no running task") {
			return err
		}
		return err
	}
	status, err := task.Status(ctx)
	switch status.Status {
	case containerd.Stopped:
		_, err := task.Delete(ctx)
		if err != nil {
			return err
		}
	case containerd.Running:
		statusC, err := task.Wait(ctx)
		if err != nil {
			c.Logger.Errorf("Container %q: error during wait: %v", container.ID(), err)
		}
		if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
			task.Delete(ctx)
			return err
		}
		status := <-statusC
		code, _, err := status.Result()
		if err != nil {
			c.Logger.Errorf("Container %q: error getting task result code: %v", container.ID(), err)
			return err
		}
		if code != 0 {
			c.Logger.Errorf("%s: exited container process: code: %d", container.ID(), status.ExitCode())
		}
		_, err = task.Delete(ctx)
		if err != nil {
			return err
		}
	case containerd.Paused:
		c.Logger.Errorf("Can't stop a paused container; unpause first")
		return err
	}
	return nil
}
