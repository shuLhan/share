// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/shuLhan/share/lib/crypto"
	libos "github.com/shuLhan/share/lib/os"
	"github.com/shuLhan/share/lib/ssh/config"
)

// Client for SSH connection.
type Client struct {
	sysEnvs map[string]string

	*ssh.Client

	config  *ssh.ClientConfig
	section *config.Section

	stdout io.Writer
	stderr io.Writer

	// identityFile that are able to connect to remoteAddr.
	identityFile string

	remoteAddr string

	listKnownHosts []string
}

// NewClientInteractive create a new SSH connection using predefined
// configuration, possibly interactively.
//
// This function may dial twice to find appropriate authentication method
// when SSH_AUTH_SOCK environment variable is set but no valid key exist and
// IdentityFile directive is specified in the Host section.
//
// If the IdentityFile is encrypted, it will prompt for passphrase in
// terminal.
//
// The following section keys are recognized and implemented by Client,
//   - Hostname
//   - IdentityAgent
//   - IdentityFile
//   - Port
//   - User
//   - UserKnownHostsFile, setting this to "none" will set HostKeyCallback
//     to [ssh.InsecureIgnoreHostKey].
func NewClientInteractive(section *config.Section) (cl *Client, err error) {
	if section == nil {
		return nil, nil
	}

	var (
		logp = `NewClientInteractive`

		sshAgent agent.ExtendedAgent
		signers  []ssh.Signer
		signer   ssh.Signer
	)

	cl = &Client{
		sysEnvs: libos.Environments(),
		config: &ssh.ClientConfig{
			User: section.User(),
		},
		section:    section,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		remoteAddr: fmt.Sprintf(`%s:%s`, section.Hostname(), section.Port()),
	}

	err = cl.setConfigHostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var sshAgentSockPath = section.IdentityAgent()
	if len(sshAgentSockPath) > 0 {
		var sshAgentSock net.Conn

		sshAgentSock, err = net.Dial("unix", sshAgentSockPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		sshAgent = agent.NewClient(sshAgentSock)

		signers, err = sshAgent.Signers()
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		signer, err = cl.dialWithSigners(signers)
		if signer != nil {
			// Client connected with one of the key in agent.
			return cl, nil
		}

		var errKey *knownhosts.KeyError
		if errors.As(err, &errKey) {
			// Host key is either unknown or mismatch with one
			// of known_hosts files, so no need to continue with
			// dialWithPrivateKeys.
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	if len(section.IdentityFile) == 0 {
		return nil, fmt.Errorf(`%s: empty IdentityFile`, logp)
	}

	err = cl.dialWithPrivateKeys(sshAgent)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return cl, nil
}

// setConfigHostKeyCallback set the config.HostKeyCallback based on the
// UserKnownHostsFile in the Section.
// If one of the UserKnownHostsFile set to "none" it will use
// [ssh.InsecureIgnoreHostKey].
func (cl *Client) setConfigHostKeyCallback() (err error) {
	var (
		logp           = `setConfigHostKeyCallback`
		userKnownHosts = cl.section.UserKnownHostsFile()

		knownHosts string
	)

	for _, knownHosts = range userKnownHosts {
		if knownHosts == config.ValueNone {
			// If one of the UserKnownHosts set to "none" always
			// accept the remote hosts.
			cl.config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
			return nil
		}

		knownHosts, err = libos.PathUnfold(knownHosts)
		if err != nil {
			return fmt.Errorf(`%s: %s: %w`, logp, knownHosts, err)
		}

		_, err = os.Stat(knownHosts)
		if err == nil {
			// Add the user known hosts file only if its exist.
			cl.listKnownHosts = append(cl.listKnownHosts, knownHosts)
		}
	}

	cl.config.HostKeyCallback, err = knownhosts.New(cl.listKnownHosts...)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// dialError return the error with clear information when the host key is
// missing or mismatch from known_hosts files.
func (cl *Client) dialError(logp string, errDial error) (err error) {
	var (
		errKey *knownhosts.KeyError
	)
	if errors.As(errDial, &errKey) {
		if len(errKey.Want) == 0 {
			err = fmt.Errorf(`%s: %w: server host key is missing from %+v`, logp, errDial, cl.listKnownHosts)
		} else {
			err = fmt.Errorf(`%s: %w: server host key mismatch in %+v`, logp, errDial, cl.listKnownHosts)
		}
	} else {
		err = fmt.Errorf(`%s: %w`, logp, errDial)
	}
	return err
}

// dialWithSigners connect to the remote machine using AuthMethod PublicKeys
// using each of signer in the list.
// On success it will return the signer that can connect to remote address.
func (cl *Client) dialWithSigners(signers []ssh.Signer) (signer ssh.Signer, err error) {
	if len(signers) == 0 {
		return nil, nil
	}
	var logp = `dialWithSigners`
	for _, signer = range signers {
		cl.config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
		cl.Client, err = ssh.Dial(`tcp`, cl.remoteAddr, cl.config)
		if err == nil {
			return signer, nil
		}
		err = cl.dialError(logp, err)
	}
	return nil, err
}

// dialWithPrivateKeys connect to the remote machine using each of the
// private key in IdentityFile.
// If the private key is encrypted it will ask for correct passphrase three
// times or continue to the next key.
// If the key is valid and sshAgent is not nil, the key will be added to the
// SSH agent.
func (cl *Client) dialWithPrivateKeys(sshAgent agent.ExtendedAgent) (err error) {
	var (
		logp = `dialWithPrivateKeys`

		identityFile string
		pkey         any
		signer       ssh.Signer
	)

	for _, identityFile = range cl.section.IdentityFile {
		fmt.Printf("%s: %s\n", logp, identityFile)

		pkey, err = crypto.LoadPrivateKeyInteractive(nil, identityFile)
		if err != nil {
			continue
		}

		signer, err = ssh.NewSignerFromKey(pkey)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}

		cl.config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}

		cl.Client, err = ssh.Dial(`tcp`, cl.remoteAddr, cl.config)
		if err == nil {
			cl.identityFile = identityFile
			break
		}
		err = cl.dialError(logp, err)
	}
	if err != nil {
		return err
	}
	if cl.Client == nil {
		// None of the private key can connect to remote address.
		return fmt.Errorf(`%s: no IdentityFile supported`, logp)
	}

	// Add key to agent.
	if sshAgent == nil {
		return nil
	}

	// TODO(ms): check for AddKeysToAgent.

	fmt.Printf("Adding key %q to agent.\n", cl.identityFile)

	var addedKey = agent.AddedKey{
		PrivateKey: pkey,
	}
	err = sshAgent.Add(addedKey)
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
	}
	return nil
}

// Close the client connection and release all resources.
func (cl *Client) Close() (err error) {
	err = cl.Client.Conn.Close()

	cl.sysEnvs = nil
	cl.Client = nil
	cl.config = nil
	cl.section = nil
	cl.stdout = nil
	cl.stderr = nil
	cl.listKnownHosts = nil

	return err
}

// Execute a command on remote server.
func (cl *Client) Execute(cmd string) (err error) {
	sess, err := cl.Client.NewSession()
	if err != nil {
		return fmt.Errorf("ssh: NewSession: " + err.Error())
	}

	sess.Stdout = cl.stdout
	sess.Stderr = cl.stderr

	for k, v := range cl.section.Environments(cl.sysEnvs) {
		err = sess.Setenv(k, v)
		if err != nil {
			log.Printf("Execute: Setenv %q=%q:%s\n", k, v, err.Error())
		}
	}

	err = sess.Run(cmd)
	if err != nil {
		err = fmt.Errorf("ssh: Run %q: %s", cmd, err.Error())
	}

	sess.Close()

	return err
}

// ScpGet copy file from remote into local storage using scp.
//
// The local file should be use the absolute path, or relative to the file in
// config.Section.WorkingDir.
func (cl *Client) ScpGet(remote, local string) (err error) {
	logp := "ScpGet"

	if len(remote) == 0 {
		return fmt.Errorf("%s: empty remote file", logp)
	}
	if len(local) == 0 {
		return fmt.Errorf("%s: empty local file", logp)
	}

	remote = fmt.Sprintf("%s@%s:%s", cl.section.User(), cl.section.Hostname(), remote)

	args := []string{`-r`, `-P`, cl.section.Port(), `-i`, cl.identityFile}
	args = append(args, remote)
	args = append(args, local)

	cmd := exec.Command("scp", args...)

	cmd.Dir = cl.section.WorkingDir
	cmd.Stdout = cl.stdout
	cmd.Stderr = cl.stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %q: %s", logp, cmd.Args, err.Error())
	}

	return nil
}

// ScpPut copy a file from local storage to remote using scp command.
//
// The local file should be use the absolute path, or relative to the file in
// config.Section's WorkingDir.
func (cl *Client) ScpPut(local, remote string) (err error) {
	logp := "ScpPut"

	if len(local) == 0 {
		return fmt.Errorf("%s: empty local file", logp)
	}
	if len(remote) == 0 {
		return fmt.Errorf("%s: empty remote file", logp)
	}

	remote = fmt.Sprintf("%s@%s:%s", cl.section.User(), cl.section.Hostname(), remote)

	args := []string{`-r`, `-P`, cl.section.Port(), `-i`, cl.identityFile}
	args = append(args, local)
	args = append(args, remote)

	cmd := exec.Command("scp", args...)

	cmd.Dir = cl.section.WorkingDir
	cmd.Stdout = cl.stdout
	cmd.Stderr = cl.stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %q: %s", logp, cmd.Args, err.Error())
	}

	return nil
}

// SetSessionOutputError set the standard output and error for future remote
// execution.
func (cl *Client) SetSessionOutputError(stdout, stderr io.Writer) {
	if stdout != nil {
		cl.stdout = stdout
	}
	if stderr != nil {
		cl.stderr = stderr
	}
}

func (cl *Client) String() string {
	return cl.section.User() + "@" + cl.section.Hostname() + ":" + cl.section.Port()
}
