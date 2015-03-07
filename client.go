package main

import (
	"encoding/json"
	"log"
	"sync"

	"centrifugo/sockjs"
	"github.com/nu7hatch/gouuid"
)

type connection interface {
	getUid() string
	getProject() string
	getUser() string
}

type client struct {
	sync.Mutex
	session         sockjs.Session
	uid             string
	project         string
	user            string
	timestamp       int
	token           string
	defaultInfo     map[string]interface{}
	channelInfo     map[string]interface{}
	isAuthenticated bool
	channels        map[string]string
	closeChannel    chan struct{}
}

func newClient(session sockjs.Session, closeChannel chan struct{}) (*client, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	return &client{
		uid:          uid.String(),
		session:      session,
		closeChannel: closeChannel,
	}, nil
}

func (c *client) getUid() string {
	return c.uid
}

func (c *client) getProject() string {
	return c.project
}

func (c *client) getUser() string {
	return c.user
}

type params map[string]interface{}

type clientCommand struct {
	Method string
	Params params
	Uid    string
}

type clientCommands []clientCommand

func getMessageType(msgBytes []byte) (string, error) {
	var f interface{}
	err := json.Unmarshal(msgBytes, &f)
	if err != nil {
		return "", err
	}
	switch f.(type) {
	case map[string]interface{}:
		return "map", nil
	case []interface{}:
		return "array", nil
	default:
		return "", ErrInvalidClientMessage
	}
}

func getCommandsFromClientMessage(msgBytes []byte, msgType string) ([]clientCommand, error) {
	var commands []clientCommand
	switch msgType {
	case "map":
		// single command request
		var command clientCommand
		err := json.Unmarshal(msgBytes, &command)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	case "array":
		// array of commands received
		err := json.Unmarshal(msgBytes, &commands)
		if err != nil {
			return nil, err
		}
	}
	return commands, nil
}

func (c *client) handleMessage(msg string) error {
	msgBytes := []byte(msg)
	msgType, err := getMessageType(msgBytes)
	if err != nil {
		return err
	}

	commands, err := getCommandsFromClientMessage(msgBytes, msgType)
	if err != nil {
		return err
	}

	err = c.handleCommands(commands)
	return err
}

func (c *client) handleCommands(commands []clientCommand) error {
	var err error
	var mr multiResponse
	for _, command := range commands {
		resp, err := c.handleCommand(command)
		if err != nil {
			return err
		}
		mr = append(mr, resp)
	}
	jsonResp, err := mr.toJson()
	if err != nil {
		return err
	}
	err = c.session.Send(string(jsonResp))
	return err
}

func (c *client) handleCommand(command clientCommand) (response, error) {
	var err error
	var resp response
	method := command.Method
	params := command.Params

	if method != "connect" && !c.isAuthenticated {
		return response{}, ErrUnauthorized
	}

	switch method {
	case "connect":
		resp, err = c.handleConnect(params)
	case "subscribe":
		resp, err = c.handleSubscribe(params)
	case "publish":
		resp, err = c.handlePublish(params)
	default:
		return response{}, ErrMethodNotFound
	}
	if err != nil {
		return response{}, err
	}

	resp.Method = method
	resp.Uid = command.Uid
	return resp, nil
}

func (c *client) handleConnect(ps params) (response, error) {
	return response{}, nil
}

func (c *client) handleSubscribe(ps params) (response, error) {
	return response{}, nil
}

func (c *client) handlePublish(ps params) (response, error) {
	return response{}, nil
}

func (c *client) printIsAuthenticated() {
	log.Println(c.isAuthenticated)
}