package containerd

import "github.com/containerd/containerd/cio"

type logURI struct {
	config cio.Config
}

func (l *logURI) Config() cio.Config {
	return l.config
}

func (l *logURI) Cancel() {

}

func (l *logURI) Wait() {

}

func (l *logURI) Close() error {
	return nil
}

func logFile(stdout string, stderr string, terminal bool) cio.Creator {
	var terminalString string
	if terminal {
		terminalString = "true"
	} else {
		terminalString = "false"
	}

	print("logURI:\n")
	print("stdout: " + stdout + "\nstderr: " + stderr + "\nterminal: " + terminalString + "\n")

	return func(_ string) (cio.IO, error) {
		return &logURI{
			config: cio.Config{
				Stdout:   stdout,
				Stderr:   stderr,
				Terminal: terminal,
			},
		}, nil
	}
}
