package server

import (
	"errors"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"pass"`
	KeyFile  string `json:"keyfile"`
}

func (server *Server) Reconnect(client **ssh.Client) error {
	if client == nil {
		return errors.New("client is nil")
	}
	if *client != nil {
		server.Disonnect(client)
	}
	c, err := server.Connect()
	if err != nil {
		return err
	}
	*client = c
	return nil
}

func (server *Server) Disonnect(client **ssh.Client) error {
	if client == nil {
		return errors.New("client is nil")
	}
	if *client == nil {
		return errors.New("client is nil")
	}
	err := (*client).Close()
	*client = nil
	return err
}

func (server *Server) Connect() (*ssh.Client, error) {
	var client *ssh.Client = nil
	var err error
	if len(server.Host) > 0 {
		var sig ssh.Signer = nil
		if len(server.KeyFile) > 0 {
			key, err := os.ReadFile(server.KeyFile)
			if err == nil {
				if len(server.Password) > 0 {
					sig, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(server.Password))
				} else {
					sig, err = ssh.ParsePrivateKey([]byte(key))
				}
			}

			if err != nil {
				return nil, err
			}
		}
		auth := []ssh.AuthMethod{}
		if sig != nil {
			auth = append(auth, ssh.PublicKeys(sig))
		} else if len(server.KeyFile) == 0 {
			auth = append(auth, ssh.Password(server.Password), ssh.KeyboardInteractive(
				func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
					answers = make([]string, len(questions))
					for n := range questions {
						answers[n] = server.Password
					}
					return answers, nil
				}))
		}

		config := &ssh.ClientConfig{
			User: server.User,
			Auth: auth,
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}

		client, err = ssh.Dial("tcp", server.Host+":"+strconv.Itoa(server.Port), config)
		if err != nil {
			return nil, err
		}
	} else {
		client = nil
	}
	return client, nil
}
