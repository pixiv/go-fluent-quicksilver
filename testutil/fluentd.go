// Copyright 2018 pixiv Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Fluentd struct {
	Pid  int
	Exit <-chan error

	buf bytes.Buffer
}

type dupWriter struct {
	writers []io.Writer
}

func (w *dupWriter) Write(p []byte) (n int, err error) {
	for _, iw := range w.writers {
		in, ie := iw.Write(p)
		n = in
		if ie != nil {
			return in, ie
		}
	}
	return
}

func SpawnFluentd(ctx context.Context, config string, args []string) (*Fluentd, error) {
	fluentd := &Fluentd{}

	w := &dupWriter{
		writers: []io.Writer{
			&fluentd.buf,
			os.Stdout,
		},
	}

	a := []string{
		"bundle", "exec", "fluentd",
		"--no-supervisor",
		"-c", "./fluent.conf",
		"-i", config,
	}
	a = append(a, args...)
	cmd := exec.CommandContext(ctx, a[0], a[1:]...)
	cmd.Dir = "../testutil"
	cmd.Stderr = os.Stderr
	cmd.Stdout = w

	ch := make(chan error)
	go func() {
		ch <- cmd.Start()
		ch <- cmd.Wait()
	}()
	err := <-ch
	if err != nil {
		return nil, err
	}
	fluentd.Pid = cmd.Process.Pid
	fluentd.Exit = ch

	return fluentd, nil
}

func (f *Fluentd) Bytes() []byte {
	return f.buf.Bytes()
}

func (f *Fluentd) Kill() error {
	p, err := os.FindProcess(f.Pid)
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGTERM)
}
