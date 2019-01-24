package server

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2"

	"github.com/gliderlabs/comlab/pkg/com"
	"github.com/gliderlabs/ssh"
	"github.com/google/go-github/github"
	"github.com/progrium/go-shell"

	tarutil "github.com/gliderlabs/exp/gitscript/lib/tar"
)

func init() {
	shell.Panic = false
}

func fatal(sess ssh.Session, err error) bool {
	if err != nil {
		fmt.Fprintln(sess.Stderr(), err.Error())
		sess.Exit(1)
		return true
	}
	return false
}

func (c *Component) HandleSSH(sess ssh.Session) {
	if len(sess.Command()) > 1 && sess.Command()[0] == "build" {
		repo := sess.Command()[1]
		repoParts := strings.Split(repo, "/")

		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: sess.User()},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)
		user, _, err := client.Users.Get(ctx, "")
		if fatal(sess, err) {
			return
		}
		perm, _, err := client.Repositories.GetPermissionLevel(ctx, repoParts[0], repoParts[1], *user.Login)
		if fatal(sess, err) {
			return
		}
		if perm.GetPermission() != "write" && perm.GetPermission() != "admin" {
			// TODO: using GetPermission requires push anyway, so this is silly
			fatal(sess, errors.New("write permission or greater is required on "+repo))
			return
		}
		//fmt.Fprintln(sess, *user.Login)

		workdir := com.GetString("workdir")
		if len(workdir) < 4 {
			panic("bad workdir")
		}
		shell.Run("rm", workdir+"/*")
		shell.Run("echo '{}' > ", shell.Path(workdir, "package.json"))

		err = tarutil.Untar(tar.NewReader(sess), workdir)
		if fatal(sess, err) {
			return
		}

		deps := strings.Split(
			shell.Run("cd", workdir, " && depcheck . --json | jq -r '.missing|keys|join(\" \")'").String(), " ")
		deps = append(deps, "brfs")
		fmt.Fprintln(sess, "Fetching dependencies...")
		shell.Run("cd", workdir, " && yarn add", strings.Join(deps, " "))
		shell.Run(`echo 'require("https");' >> `, shell.Path(workdir, "node_modules/github/lib/index.js"))
		fmt.Fprintln(sess, "Building bundle...")
		jsFiles := strings.Split(
			shell.Run("cd", workdir, " && ls *.js").String(), "\n")
		bundle := shell.Run("cd", workdir, " && browserify -g brfs", strings.Join(jsFiles, " ")).Bytes()
		fmt.Fprintln(sess, "Bundle size:", len(bundle))

	} else {
		fmt.Fprintf(sess.Stderr(), "%s", sess.Command())
	}
}

func (c *Component) HandleAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}
