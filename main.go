package main

import (
	"fmt"
	"io"
	"net"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"
)

// LoggerWrapper provides thread-safe logging for the GUI
type LoggerWrapper struct {
	LogBox *widget.Entry
	Mutex  sync.Mutex
}

func (lw *LoggerWrapper) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	lw.Mutex.Lock()
	defer lw.Mutex.Unlock()
	lw.LogBox.SetText(lw.LogBox.Text + message + "\n")
}

func main() {
	// Create a Fyne app
	myApp := app.New()
	myWindow := myApp.NewWindow("SSH Reverse Tunnel")

	// Define input fields for SSH configuration
	sshServerEntry := widget.NewEntry()
	sshServerEntry.SetPlaceHolder("AWS Server Address (e.g., aws-server:22)")

	sshUserEntry := widget.NewEntry()
	sshUserEntry.SetPlaceHolder("SSH Username")

	sshPasswordEntry := widget.NewPasswordEntry()
	sshPasswordEntry.SetPlaceHolder("SSH Password")

	localServerEntry := widget.NewEntry()
	localServerEntry.SetText("localhost:3000")
	localServerEntry.SetPlaceHolder("Local Server Address (e.g., localhost:3000)")

	remotePortEntry := widget.NewEntry()
	remotePortEntry.SetText("0.0.0.0:8080")
	remotePortEntry.SetPlaceHolder("Remote Port (e.g., 0.0.0.0:8080)")

	// Log display box
	logBox := widget.NewMultiLineEntry()
	logBox.SetPlaceHolder("Logs will appear here...")
	logBox.Disable()

	// Logger wrapper for safe log printing
	logger := &LoggerWrapper{LogBox: logBox}

	// Start button action
	startButton := widget.NewButton("Start Tunnel", func() {
		sshServer := sshServerEntry.Text
		sshUser := sshUserEntry.Text
		sshPassword := sshPasswordEntry.Text
		localServer := localServerEntry.Text
		remotePort := remotePortEntry.Text

		go startTunnel(sshServer, sshUser, sshPassword, localServer, remotePort, logger)
	})

	// Layout the form and log box
	form := container.NewVBox(
		widget.NewLabel("Configure SSH Tunnel"),
		sshServerEntry,
		sshUserEntry,
		sshPasswordEntry,
		localServerEntry,
		remotePortEntry,
		startButton,
		widget.NewLabel("Logs"),
		logBox,
	)

	myWindow.SetContent(form)
	myWindow.Resize(fyne.NewSize(500, 400))
	myWindow.ShowAndRun()
}

func startTunnel(sshServer, sshUser, sshPassword, localServer, remotePort string, logger *LoggerWrapper) {
	// SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword), // Use ssh.PublicKeys() for private key authentication
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Skip host key verification (for testing only)
	}

	logger.Printf("Connecting to SSH server: %s\n", sshServer)

	// Connect to the SSH server
	sshConn, err := ssh.Dial("tcp", sshServer, sshConfig)
	if err != nil {
		logger.Printf("Failed to connect to SSH server: %v\n", err)
		return
	}
	defer sshConn.Close()
	logger.Printf("Connected to SSH server: %s\n", sshServer)

	// Listen on the remote server for forwarded connections
	remoteListener, err := sshConn.Listen("tcp", remotePort)
	if err != nil {
		logger.Printf("Failed to start remote listener: %v\n", err)
		return
	}
	defer remoteListener.Close()
	logger.Printf("Listening on remote port %s for incoming connections...\n", remotePort)

	for {
		// Accept connections coming to the remote server
		remoteConn, err := remoteListener.Accept()
		if err != nil {
			logger.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		logger.Printf("Accepted connection from remote client.\n")

		// Connect to the local server
		localConn, err := net.Dial("tcp", localServer)
		if err != nil {
			logger.Printf("Failed to connect to local server: %v\n", err)
			remoteConn.Close()
			continue
		}

		// Forward traffic between remote and local connections
		go func() {
			defer remoteConn.Close()
			defer localConn.Close()
			io.Copy(remoteConn, localConn)
		}()
		go func() {
			defer remoteConn.Close()
			defer localConn.Close()
			io.Copy(localConn, remoteConn)
		}()
	}
}
