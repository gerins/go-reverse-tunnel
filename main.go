package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"
)

// LoggerWrapper provides thread-safe logging for the GUI
type LoggerWrapper struct {
	LogBox    *widget.Entry
	LogScroll *container.Scroll
	Mutex     sync.Mutex
}

func (lw *LoggerWrapper) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	lw.Mutex.Lock()
	defer lw.Mutex.Unlock()

	withTime := fmt.Sprintf("%v | %v\n", time.Now().Format("15:04:05"), message)
	lw.LogBox.SetText(withTime + lw.LogBox.Text)
}

type Config struct {
	RemoteAddress string `json:"remote_address"`
	User          string `json:"username"`
	Password      string `json:"password"`
	RemotePort    string `json:"remote_port"`
	LocalAddress  string `json:"local_address"`
}

func (c *Config) LoadFromFile() {
	fileName := fmt.Sprintf("%v.txt", "go_tunnel_config")
	fileLocation := fmt.Sprintf("%v/%v", os.TempDir(), fileName)
	file, err := os.Open(fileLocation)
	if err != nil {
		return
	}

	fileContent, _ := io.ReadAll(file)

	if len(fileContent) > 0 {
		json.Unmarshal(fileContent, c)
	}
}

func (c *Config) SaveToFile() error {
	fileName := fmt.Sprintf("%v.txt", "go_tunnel_config")
	fileLocation := fmt.Sprintf("%v/%v", os.TempDir(), fileName)
	file, err := os.Create(fileLocation)
	if err != nil {
		return err
	}

	configJSON, _ := json.Marshal(c)
	file.Write(configJSON)
	file.Close()

	return nil
}

func main() {
	config := Config{}
	config.LoadFromFile()

	// Create a Fyne app
	myApp := app.New()
	myWindow := myApp.NewWindow("SSH Reverse Tunnel")

	// Define input fields for SSH configuration
	sshServerEntry := widget.NewEntry()
	sshServerEntry.SetPlaceHolder("Remote Server Address (e.g., aws-server:22)")
	if config.RemoteAddress != "" {
		sshServerEntry.SetText(config.RemoteAddress)
	}

	sshUserEntry := widget.NewEntry()
	sshUserEntry.SetPlaceHolder("SSH Username")
	if config.User != "" {
		sshUserEntry.SetText(config.User)
	}

	sshPasswordEntry := widget.NewPasswordEntry()
	sshPasswordEntry.SetPlaceHolder("SSH Password")
	if config.Password != "" {
		sshPasswordEntry.SetText(config.Password)
	}

	remotePortEntry := widget.NewEntry()
	remotePortEntry.SetPlaceHolder("Remote Port (e.g., 3000)")
	if config.RemotePort != "" {
		remotePortEntry.SetText(config.RemotePort)
	}

	localServerEntry := widget.NewEntry()
	localServerEntry.SetPlaceHolder("Local Server Address (e.g., localhost:8080)")
	if config.LocalAddress != "" {
		localServerEntry.SetText(config.LocalAddress)
	}

	// Log display box
	logBox := widget.NewMultiLineEntry()
	logBox.SetPlaceHolder("Logs will appear here...")
	logBox.Disable()

	// Wrap logBox in a scrollable container
	logContainer := container.NewVScroll(logBox)
	logContainer.SetMinSize(fyne.NewSize(480, 200)) // Set larger log box size

	// Logger wrapper for safe log printing
	logger := &LoggerWrapper{LogBox: logBox, LogScroll: logContainer}

	// Start button action
	startButton := widget.NewButton("Start Tunnel", func() {
		config = Config{
			RemoteAddress: sshServerEntry.Text,
			User:          sshUserEntry.Text,
			Password:      sshPasswordEntry.Text,
			RemotePort:    remotePortEntry.Text,
			LocalAddress:  localServerEntry.Text,
		}

		if err := config.SaveToFile(); err != nil {
			logger.Printf("failed saving config to file temp, %v", err)
		}

		go startTunnel(config, logger)
	})

	// Layout the form and log box
	form := container.NewVBox(
		widget.NewLabel("Configure SSH Tunnel"),
		sshServerEntry,
		sshUserEntry,
		sshPasswordEntry,
		remotePortEntry,
		localServerEntry,
		startButton,
		// widget.NewLabel("Logs"),
		logContainer,
	)

	myWindow.SetContent(form)
	myWindow.Resize(fyne.NewSize(500, 500)) // Resize the window to accommodate the larger log box
	myWindow.ShowAndRun()
}

func startTunnel(config Config, logger *LoggerWrapper) {
	// SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password), // Use ssh.PublicKeys() for private key authentication
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Skip host key verification (for testing only)
	}

	logger.Printf("Connecting to SSH server: %s", config.RemoteAddress)

	// Connect to the SSH server
	sshConn, err := ssh.Dial("tcp", config.RemoteAddress, sshConfig)
	if err != nil {
		logger.Printf("Failed to connect to SSH server: %v", err)
		return
	}
	defer sshConn.Close()
	logger.Printf("Connected to SSH server: %s", config.RemoteAddress)

	// Listen on the remote server for forwarded connections
	remotePort := fmt.Sprintf("0.0.0.0:%v", config.RemotePort)
	remoteListener, err := sshConn.Listen("tcp", remotePort)
	if err != nil {
		logger.Printf("Failed to start remote listener: %v", err)
		return
	}
	defer remoteListener.Close()
	logger.Printf("Listening on remote port %s for incoming connections...", remotePort)

	for {
		// Accept connections coming to the remote server
		remoteConn, err := remoteListener.Accept()
		if err != nil {
			logger.Printf("Failed to accept connection: %v", err)
			continue
		}
		logger.Printf("Accepted connection from remote client.")

		// Connect to the local server
		localConn, err := net.Dial("tcp", config.LocalAddress)
		if err != nil {
			logger.Printf("Failed to connect to local server: %v", err)
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
