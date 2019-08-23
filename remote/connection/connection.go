package connection

import (
	"fmt"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

type Connection struct {
	SSH *SSHConnection
}

type SSHConnection struct {
	Host string
	User string
	Sudo bool
	Port int
}

func Connect(connString string) (*Session, error) {
	conn, err := DecodeConnection(connString)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode conn: %s", err)
	}
	sess, err := conn.Handle().Connect()
	if err != nil {
		return nil, fmt.Errorf("Cannot connect: %s", err)
	}
	return sess, nil
}

func DecodeConnection(s string) (*Connection, error) {
	var c Connection
	err := json.Unmarshal([]byte(s), &c)
	return &c, err
}

func (c *Connection) Encode() (encoded string, id string) {
	bytes, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	checksum := sha1.Sum(bytes)
	return string(bytes), hex.EncodeToString(checksum[:])
}

func (c *Connection) Id() string {
	_, id := c.Encode()
	return id
}

func (c *Connection) Handle() *Handle {
	id := c.Id()
	h := Handles[id]
	if h == nil {
		h = NewHandle(c)
		Handles[id] = h
	}
	return h
}
