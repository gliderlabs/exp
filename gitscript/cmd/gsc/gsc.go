package main

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"

	tarutil "github.com/gliderlabs/exp/gitscript/lib/tar"
)

func fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func PrivateKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func main() {
	out, err := exec.Command("git", strings.Split("config --get remote.origin.url", " ")...).Output()
	fatal(err)
	if !strings.Contains(string(out), "github.com") {
		fatal(errors.New("not in a github repo"))
	}
	parts := strings.Split(string(out), ":")
	if len(parts) < 2 {
		fatal(errors.New("unrecognized origin remote"))
	}
	repo := strings.TrimSuffix(parts[1], ".git\n")
	fmt.Println("* Using repository", repo)

	auth, err := PrivateKeyFile("/Users/progrium/.ssh/id_rsa")
	fatal(err)
	config := &ssh.ClientConfig{
		User: os.Getenv("GITSCRIPT_TOKEN"),
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", "localhost:2222", config)
	fatal(err)
	defer conn.Close()

	sess, err := conn.NewSession()
	fatal(err)
	stdout, err := sess.StdoutPipe()
	fatal(err)
	stderr, err := sess.StderrPipe()
	fatal(err)
	stdin, err := sess.StdinPipe()
	fatal(err)

	fatal(sess.Start(fmt.Sprintf("build %s", repo)))
	defer sess.Close()

	go io.Copy(os.Stderr, stderr)
	go func() {
		tarWriter := tar.NewWriter(stdin)
		defer tarWriter.Close()
		fatal(tarutil.Tarball([]string{"."}, tarWriter))
	}()
	io.Copy(os.Stdout, stdout)
}
