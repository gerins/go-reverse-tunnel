
# Reverse SSH Tunnel with Golang for Local API Forwarding

This project demonstrates how to use **Go** to create an **SSH reverse tunnel** to forward traffic from a remote server (e.g., an AWS EC2 instance) to a local server running on your machine. This is particularly useful for exposing local REST APIs to external clients securely.


## Features

- Establishes an SSH connection to a remote server.
- Forwards traffic received on a remote port to your local server.
- Supports **HTTP REST APIs** or any TCP-based protocols.
- Seamlessly handles requests and responses over the tunnel.


## Use Case

You can expose your local development server (e.g., `localhost:3000`) to the internet by forwarding traffic from an AWS server.

For example:

1. A client sends a request to `http://aws-server-address:8080/callback`.
2. The Go program forwards the request to your local server at `http://localhost:3000/callback`.
3. Your local server processes the request and sends the response back to the client.


## Prerequisites

- Go 1.18+ installed on your machine.
- An SSH-accessible remote server (e.g., AWS EC2).
- Local server running on your machine (e.g., `localhost:3000`).


## Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/yourusername/reverse-ssh-tunnel.git
   cd reverse-ssh-tunnel
   ```

2. **Install dependencies**:

   The project uses the `golang.org/x/crypto/ssh` package.

   ```bash
   go get golang.org/x/crypto/ssh
   ```

3. **Run your local server** (example for Go):

   ```go
   package main

   import (
       "fmt"
       "net/http"
   )

   func callbackHandler(w http.ResponseWriter, r *http.Request) {
       fmt.Println("Received callback request")
       fmt.Fprintf(w, "Callback received successfully!")
   }

   func main() {
       http.HandleFunc("/callback", callbackHandler)
       fmt.Println("Local server running on http://localhost:3000")
       http.ListenAndServe(":3000", nil)
   }
   ```

   Start the local server:

   ```bash
   go run local_server.go
   ```


## Usage

### Update the Go Program

Modify the `main.go` file with your configuration:

- Replace `aws-server-address` with the address of your remote server.
- Update the SSH username (`sshUser`) and password (`sshPassword`).
- Set the local server (`localhost:3000`) and remote port (`8080`).


### Run the Reverse SSH Tunnel

Start the Go program:

```bash
go run main.go
```

The program will:

1. Connect to your remote SSH server.
2. Listen for incoming traffic on `0.0.0.0:8080` (AWS server).
3. Forward requests to `localhost:3000`.


### Test the Forwarding

Send a request to the remote server:

```bash
curl http://aws-server-address:8080/callback
```

The local server will handle the request and respond:

```
Callback received successfully!
```

You will also see the output on your local server console:

```
Received callback request
```


## File Structure

```
reverse-ssh-tunnel/
│
├── main.go             # Reverse SSH tunnel code
├── local_server.go     # Example local server handling "/callback"
└── README.md           # Project documentation
```



## Security Notes

- Replace password-based authentication with **SSH private key authentication** for production environments.
- Make sure the remote server's firewall/security group allows inbound traffic on the forwarded port (e.g., `8080`).



## Dependencies

- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh)


## License

This project is licensed under the MIT License.


## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you'd like to change.


## Author

**Your Name**  
[GitHub](https://github.com/gerins) | [LinkedIn](#)
