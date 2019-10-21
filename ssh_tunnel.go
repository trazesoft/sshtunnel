package sshtunnel

import (
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	// "time"
	"strconv"
)

type SSHTunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint
	Config *ssh.ClientConfig
	Log    *log.Logger
}

func (tunnel *SSHTunnel) logf(fmt string, args ...interface{}) {
	if tunnel.Log != nil {
		tunnel.Log.Printf(fmt, args...)
	}
}

func (tunnel *SSHTunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	tunnel.Local.Port = listener.Addr().(*net.TCPAddr).Port

	go createSshConnection(tunnel)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		tunnel.logf("accepted connection")
		tunnel.logf("local port %s", strconv.Itoa(conn.RemoteAddr().(*net.TCPAddr).Port))
		go tunnel.forward(conn)
	}
}

var sshConnectionWorking bool = false
var workingFlag chan bool = make(chan bool)
// var done bool = false
// var total int = 0
var serverConn *ssh.Client

	
func createSshConnection(tunnel *SSHTunnel){

	var serverErr interface{}
	var retry int

	if sshConnectionWorking == false {

		sshConnectionWorking = true

		// createSshConnection(tunnel)
		for {

			retry ++
			// c, err := d.dialDNS(ctx, network, server) 
			serverConn, serverErr = ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
			if serverErr != nil {
				tunnel.logf("server dial error: %s", serverErr)
			} else {
				close(workingFlag)
				retry = 0
				workingFlag = make(chan bool)
				sshConnectionWorking = false
		// tunnel.logf("Loaded")
				//time.Sleep(10 * time.Second)
				break
			}

		}
	} else if serverConn == nil {
				
			
		tunnel.logf("Waiting for the bus!")
		<-workingFlag
		tunnel.logf("The bus has arrived!")

	}

				
		tunnel.logf("Loaded finished")
}
func (tunnel *SSHTunnel) forward(localConn net.Conn) {

	var retry int


	/*if sshConnectionWorking == false {

		sshConnectionWorking = true

		createSshConnection(tunnel)
	} else if serverConn == nil {
				
			
		tunnel.logf("Waiting for the bus!")
		<-workingFlag
		tunnel.logf("The bus has arrived!")

	}
	// serverConn.Status()

	if retry > 1 {
		tunnel.logf("Retry server: %s", strconv.Itoa(retry))
	}*/

	// tunnel.logf("connected to %s (1 of 2)\n", tunnel.Server.String())

	var remoteConn net.Conn
	var remoteError interface{}
	retry = 0

	for {
		retry ++
		remoteConn, remoteError = serverConn.Dial("tcp", tunnel.Remote.String())
		// remoteConn, remoteError = serverConn.Dial("tcp", "127.0.0.1:3306")
		if remoteError != nil {
			if retry > 3 {
				//serverConn = nil
				createSshConnection(tunnel)
			}
			tunnel.logf("remote dial error: %s", remoteError)
		}else{

			break
		}
	}
	if retry > 1 {
		tunnel.logf("Retry remote: %s", strconv.Itoa(retry))
	}

	tunnel.logf("connected to %s (2 of 2)\n", tunnel.Remote.String())

	go func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()
		_, err := io.Copy(writer, reader)
		if err != nil {
			tunnel.logf("io.Copy local to remote warm: %s", err)
		}
	}(localConn, remoteConn)

	go func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()
		_, err := io.Copy(writer, reader)
		if err != nil {
			tunnel.logf("io.Copy remote to local warm: %s", err)
		}
	}(remoteConn, localConn)

}

func NewSSHTunnel(tunnel string, auth ssh.AuthMethod, destination string, localPort string) *SSHTunnel {
	// A random port will be chosen for us.
	localEndpoint := NewEndpoint("localhost:" + localPort)

	server := NewEndpoint(tunnel)
	if server.Port == 0 {
		server.Port = 22
	}

	sshTunnel := &SSHTunnel{
		Config: &ssh.ClientConfig{
			User: server.User,
			Auth: []ssh.AuthMethod{auth},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// Always accept key.
				return nil
			},
		},
		Local:  localEndpoint,
		Server: server,
		Remote: NewEndpoint(destination),
	}

	return sshTunnel
}


