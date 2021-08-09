// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package chirp implements a client to communicate with BIRD.
package chirp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// New creates a BIRDClient.
func New(socket string) (*BIRDClient, error) {
	conn, err := newConn(socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BIRD: %w", err)
	}
	return &BIRDClient{
		socket:  socket,
		conn:    conn,
		scanner: bufio.NewScanner(conn),
	}, nil
}

// BIRDClient handles communication with BIRD.
type BIRDClient struct {
	mu      sync.Mutex
	socket  string
	conn    net.Conn
	scanner *bufio.Scanner
}

// DisableProtocol disables the provided protocol.
func (b *BIRDClient) DisableProtocol(protocol string) error {
	out, err := b.exec(fmt.Sprintf("disable %s\n", protocol))
	if err != nil {
		return err
	}
	if strings.Contains(out, fmt.Sprintf("%s: already disabled", protocol)) {
		return nil
	} else if strings.Contains(out, fmt.Sprintf("%s: disabled", protocol)) {
		return nil
	}
	return fmt.Errorf("failed to disable %s: %v", protocol, out)
}

// EnableProtocol enables the provided protocol.
func (b *BIRDClient) EnableProtocol(protocol string) error {
	out, err := b.exec(fmt.Sprintf("enable %s\n", protocol))
	if err != nil {
		return err
	}
	if strings.Contains(out, fmt.Sprintf("%s: already enabled", protocol)) {
		return nil
	} else if strings.Contains(out, fmt.Sprintf("%s: enabled", protocol)) {
		return nil
	}
	return fmt.Errorf("failed to enable %s: %v", protocol, out)
}

func (b *BIRDClient) exec(cmd string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, err := b.conn.Write([]byte(cmd)); err != nil {
		return "", err
	}
	if !b.scanner.Scan() {
		return "", fmt.Errorf("reading response from bird failed")
	}
	if err := b.scanner.Err(); err != nil {
		return "", err
	}
	return b.scanner.Text(), nil
}

func newConn(path string) (net.Conn, error) {
	c, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	if _, err := c.Read(make([]byte, 4096)); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}
