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

// SendMessage sends a message with 8-byte string length prefix to the server
func (ch *CommunicationHandler) SendMessage(msg string) error {
	// Create a single buffer with 8-byte string length prefix + message
	lengthStr := fmt.Sprintf("%08d", len(msg))  // Format as 8-character string (00000000-99999999)
	
	// Combine length string and message into one buffer
	buffer := append([]byte(lengthStr), []byte(msg)...)
	
	writtenBytes := 0
	for writtenBytes < len(buffer) {
		n, err := ch.conn.Write(buffer[writtenBytes:])
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		writtenBytes += n
	}

	return nil
}

// ParseResponse parses a simple response from the server
func ParseResponse(response string) (int, error) {
	if len(response) == 0 {
		return -1, fmt.Errorf("empty response from server")
	}
	
	// The response is a number in string format
	var numericResponse int
	_, err := fmt.Sscanf(response, "%d", &numericResponse)
	if err != nil {
		return -1, fmt.Errorf("invalid response format: %w", err)
	}

	if numericResponse < 0 {
		return numericResponse, fmt.Errorf("unexpected behavior in server: %w", err) // Negative response indicates failure
	}

	return numericResponse, nil
}

// ReceiveMessage receives a simple string message from the server
func (ch *CommunicationHandler) ReceiveMessage() (int, error) {
	// Read response using bufio for simple line-based protocol
	reader := bufio.NewReader(ch.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return -1, fmt.Errorf("failed to receive message: %w", err)
	}
	
	// Remove trailing newline
	response = response[:len(response)-1]

	parsedResponse, err := ParseResponse(response)
	
	return parsedResponse, err
}

// Close closes the underlying connection
func (ch *CommunicationHandler) Close() error {
	if ch.conn != nil {
		return ch.conn.Close()
	}
	return nil
}
