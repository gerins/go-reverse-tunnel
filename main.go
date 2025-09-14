package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func logf(format string, args ...interface{}) {
	fmt.Printf("%s | %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func publicKeyAuth(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func startTunnel(remoteAddr, user, password, keyPath, remotePort, localAddr string) {
	var authMethod ssh.AuthMethod
	var err error

	if keyPath != "" {
		authMethod, err = publicKeyAuth(keyPath)
		if err != nil {
			logf("Failed loading private key: %v", err)
			return
		}
		logf("Using public key authentication")
	} else {
		authMethod = ssh.Password(password)
		logf("Using password authentication")
	}

	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	logf("Connecting to SSH server: %s", remoteAddr)
	sshConn, err := ssh.Dial("tcp", remoteAddr, sshConfig)
	if err != nil {
		logf("SSH connection error: %v", err)
		return
	}
	defer sshConn.Close()
	logf("Connected to SSH server")

	remoteBind := fmt.Sprintf("0.0.0.0:%s", remotePort)
	remoteListener, err := sshConn.Listen("tcp", remoteBind)
	if err != nil {
		logf("Remote listener error: %v", err)
		return
	}
	defer remoteListener.Close()
	logf("Listening on remote port %s", remoteBind)

	for {
		remoteConn, err := remoteListener.Accept()
		if err != nil {
			logf("Accept failed: %v", err)
			continue
		}
		logf("Remote client connected")

		localConn, err := net.Dial("tcp", localAddr)
		if err != nil {
			logf("Local connection failed: %v", err)
			remoteConn.Close()
			continue
		}

		go proxy(remoteConn, localConn)
		go proxy(localConn, remoteConn)
	}
}

func proxy(dst net.Conn, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func main() {
	remoteAddr := flag.String("remote", "", "Remote SSH server (e.g., host:22)")
	user := flag.String("user", "", "SSH username")
	password := flag.String("pass", "", "SSH password (optional if key is used)")
	keyPath := flag.String("key", "", "Path to private key file (optional if password is used)")
	remotePort := flag.String("rport", "", "Remote port to bind (e.g., 3000)")
	localAddr := flag.String("laddr", "", "Local server to forward to (e.g., localhost:8080)")

	flag.Parse()

	if *remoteAddr == "" || *user == "" || *remotePort == "" || *localAddr == "" {
		fmt.Println("Missing required arguments. Use -h for help.")
		os.Exit(1)
	}

	if *password == "" && *keyPath == "" {
		fmt.Println("Provide either -pass or -key for authentication.")
		os.Exit(1)
	}

	startTunnel(*remoteAddr, *user, *password, *keyPath, *remotePort, *localAddr)
}
