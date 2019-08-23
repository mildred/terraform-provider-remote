package connection

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

type Session struct {
	ssh    io.Closer
	Stdin  io.Writer
	Stdout io.Reader
	out    *bufio.Scanner
	End    <-chan error
}

func (s *Session) Close() error {
	return s.ssh.Close()
}

func (s *Session) runCmd(cmd string) ([]byte, []byte, error) {
	if s.out == nil {
		s.out = bufio.NewScanner(s.Stdout)
	}

	resIn := make(chan error, 1)
	go func() {
		_, err := fmt.Fprintf(s.Stdin, `
		(
			(
				( %s ) > >(base64 | sed s/^/out:/; echo)
			) 2> >(base64 | sed s/^/err:/; echo)
		)
		`, cmd)
		resIn <- err
	}()

	resOut := make(chan struct {
		out []byte
		err []byte
		er  error
	}, 1)
	go func() {
		var err error
		var stdout, stderr []byte
		defer func() {
			resOut <- struct {
				out []byte
				err []byte
				er  error
			}{stdout, stderr, err}
		}()
		stop := 0
		for stop < 2 && s.out.Scan() {
			line := s.out.Text()
			if strings.HasPrefix(line, "out:") {
				stdout = append(stdout, []byte(line[4:])...)
			} else if strings.HasPrefix(line, "err:") {
				stderr = append(stderr, []byte(line[4:])...)
			} else if line == "" {
				stop = stop + 1
			} else {
				err = fmt.Errorf("Unknown line: %s", line)
				return
			}
		}
		err = s.out.Err()
		if err != nil {
			return
		}

		stdout, err = base64.StdEncoding.DecodeString(string(stdout))
		if err != nil {
			err = fmt.Errorf("Decoding stdout: %s", err)
			return
		}

		stderr, err = base64.StdEncoding.DecodeString(string(stderr))
		if err != nil {
			err = fmt.Errorf("Decoding stderr: %s", err)
			return
		}
	}()

	for {
		select {
		case err := <-resIn:
			if err == nil {
				continue
			}
			return nil, nil, fmt.Errorf("Could not send command: %v", err)
		case res := <-resOut:
			return res.out, res.err, res.er
		case err := <-s.End:
			return nil, nil, fmt.Errorf("Interpreter failed: %v", err)
		}
	}
}

func shEscape(s string) string {
	return strings.Replace(s, `'`, `'"'"'`, -1)
}

func (s *Session) ReadFile(fname string) ([]byte, error) {
	out, err, er := s.runCmd(fmt.Sprintf("cat -- %v", shEscape(fname)))
	if er != nil {
		return nil, er
	} else if len(err) != 0 {
		return nil, fmt.Errorf("reading file: %v", err)
	} else {
		return out, nil
	}
}

func (s *Session) WriteFile(fname string, content []byte, mode os.FileMode) error {
	_, err, er := s.runCmd(fmt.Sprintf("base64 -d >%s <<EOF\n%s\nEOF\nchmod %o -- %s",
		shEscape(fname),
		base64.StdEncoding.EncodeToString(content),
		mode,
		shEscape(fname)))
	if er != nil {
		return er
	} else if len(err) != 0 {
		return fmt.Errorf("writing file: %v", err)
	} else {
		return nil
	}
}

func (s *Session) RemoveFile(fname string, force, recursive bool) error {
	flags := ""
	if force {
		flags += " -f"
	}
	if recursive {
		flags += " -r"
	}

	_, err, er := s.runCmd(fmt.Sprintf("rm %s -- %s",
		flags,
		shEscape(fname)))
	if er != nil {
		return er
	} else if len(err) != 0 {
		return fmt.Errorf("removing file: %v", err)
	} else {
		return nil
	}
}
