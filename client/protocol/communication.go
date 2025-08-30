package protocol

import (
	"bufio"
	"fmt"
	"net"
)

// CommunicationHandler handles simple string-based communication with length prefixing
type CommunicationHandler struct {
	conn net.Conn
}

// NewCommunicationHandler creates a new communication handler
func NewCommunicationHandler(conn net.Conn) *CommunicationHandler {
	return &CommunicationHandler{
		conn: conn,
	}
}

// SendMessage sends a message with 1-byte length prefix to the server
func (ch *CommunicationHandler) SendMessage(msg string) error {
	// Create a single buffer with length byte + message
	length := uint8(len(msg))
	
	// Combine length byte and message into one buffer
	buffer := append([]byte{length}, []byte(msg)...)
	
	written_bytes := 0
	for written_bytes < len(buffer) {
		n, err := ch.conn.Write(buffer[written_bytes:])
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		written_bytes += n
	}

	return nil
}

// ReceiveMessage receives a simple string message from the server
func (ch *CommunicationHandler) ReceiveMessage() (string, error) {
	// Read response using bufio for simple line-based protocol
	reader := bufio.NewReader(ch.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to receive message: %w", err)
	}
	
	// Remove trailing newline
	response = response[:len(response)-1]
	
	return response, nil
}

// Close closes the underlying connection
func (ch *CommunicationHandler) Close() error {
	if ch.conn != nil {
		return ch.conn.Close()
	}
	return nil
}
