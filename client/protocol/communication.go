package protocol

import (
	"fmt"
	"net"
	"strings"
	"strconv"
)

// Protocol constants
const (
	MessageTypeSize = 1  // Size of message type in bytes
	MessageSizeSize = 8  // Size of message size field in bytes
	MaxMessageSize  = 8 * 1024 // Maximum message size (8KB)
)

// CommunicationHandler handles communication with the new protocol format
type CommunicationHandler struct {
	conn net.Conn
}

// NewCommunicationHandler creates a new communication handler
func NewCommunicationHandler(conn net.Conn) *CommunicationHandler {
	if conn == nil {
		return nil
	}
	return &CommunicationHandler{
		conn: conn,
	}
}

// SendBatchBets sends a batch of bets to the server
func (ch *CommunicationHandler) SendBatchBets(content string) error {
	return ch.SendMessage(MessageTypeBatchBets, content)
}

// SendNotification sends a completion notification to the server
func (ch *CommunicationHandler) SendNotification(agencyID int) error {
	if agencyID <= 0 {
		return fmt.Errorf("invalid agency ID: %d", agencyID)
	}
	notification := fmt.Sprintf("%d", agencyID)
	return ch.SendMessage(MessageTypeNotification, notification)
}

// SendWinnerQuery sends a winner query to the server
func (ch *CommunicationHandler) SendWinnerQuery(agencyID int) error {
	if agencyID <= 0 {
		return fmt.Errorf("invalid agency ID: %d", agencyID)
	}
	query := fmt.Sprintf("%d", agencyID)
	return ch.SendMessage(MessageTypeWinnerQuery, query)
}

// SendMessage sends a message with the new protocol format: <type><size><content>
func (ch *CommunicationHandler) SendMessage(msgType byte, content string) error {
	if ch.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	
	if len(content) > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", len(content), MaxMessageSize)
	}
	
	// Format: <type><size_8_bytes><content>
	sizeStr := fmt.Sprintf("%08d", len(content))
	fullMessage := string(msgType) + sizeStr + content
	
	buffer := []byte(fullMessage)
	
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

// ReceiveBatchResponse receives a batch bet response from the server
// Expects message type 'R' and returns the parsed response count
func (ch *CommunicationHandler) ReceiveBatchResponse() (int, error) {
	msgType, content, err := ch.Read()
	if err != nil {
		return -1, fmt.Errorf("failed to receive batch response: %w", err)
	}
	
	// Expect response type 'R' for batch bet responses
	if msgType != 'R' {
		return -1, fmt.Errorf("unexpected message type: %c, expected 'R'", msgType)
	}
	
	parsedResponse, err := ParseResponse(content)
	return parsedResponse, err
}

// ReceiveNotificationResponse receives a notification acknowledgment from the server
// Expects message type 'A' and returns the acknowledgment status
func (ch *CommunicationHandler) ReceiveNotificationResponse() (bool, error) {
	msgType, content, err := ch.Read()
	if err != nil {
		return false, fmt.Errorf("failed to receive notification response: %w", err)
	}
	
	// Expect acknowledgment type 'A' for notification responses
	if msgType != 'A' {
		return false, fmt.Errorf("unexpected message type: %c, expected 'A'", msgType)
	}
	
	// Parse the acknowledgment (1 for success, 0 for failure)
	if content == "1" {
		return true, nil
	} else if content == "0" {
		return false, nil
	} else {
		return false, fmt.Errorf("invalid acknowledgment content: %s", content)
	}
}

// ReceiveWinnerList receives a winner list from the server
// Expects message type 'W' for winner list or 'T' for waiting response
func (ch *CommunicationHandler) ReceiveWinnerList() ([]string, error) {
	msgType, content, err := ch.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to receive winner list: %w", err)
	}
	
	// Check message type and handle accordingly
	switch msgType {
	case 'W': // Winner list
		// Parse winner list
		if content == "" {
			return []string{}, nil // No winners
		}
		winners := strings.Split(content, ",")
		return winners, nil
		
	case 'T': // Waiting response
		return nil, fmt.Errorf("server waiting for other clients to complete")
		
	default:
		return nil, fmt.Errorf("unexpected message type: %c, expected 'W' or 'T'", msgType)
	}
}

// Read receives a message using the new protocol format: <type><size><content>
func (ch *CommunicationHandler) Read() (byte, string, error) {
	if ch.conn == nil {
		return 0, "", fmt.Errorf("connection is nil")
	}
	
	// First read the message type (1 byte)
	typeBuffer := make([]byte, MessageTypeSize)
	readBytes := 0
	for readBytes < MessageTypeSize {
		n, err := ch.conn.Read(typeBuffer[readBytes:])
		if err != nil {
			return 0, "", fmt.Errorf("failed to read message type: %w", err)
		}
		readBytes += n
	}
	msgType := typeBuffer[0]
	
	// Then read the message size (8 characters for 8-digit length)
	sizeBuffer := make([]byte, MessageSizeSize)
	readBytes = 0
	for readBytes < MessageSizeSize {
		n, err := ch.conn.Read(sizeBuffer[readBytes:])
		if err != nil {
			return 0, "", fmt.Errorf("failed to read message size: %w", err)
		}
		readBytes += n
	}
	
	totalSize, err := strconv.Atoi(string(sizeBuffer))
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse message size: %w", err)
	}
	
	if totalSize < 0 || totalSize > MaxMessageSize {
		return 0, "", fmt.Errorf("invalid message size: %d (max %d)", totalSize, MaxMessageSize)
	}
	
	// Read the complete message content
	contentBuffer := make([]byte, totalSize)
	readBytes = 0
	for readBytes < totalSize {
		n, err := ch.conn.Read(contentBuffer[readBytes:])
		if err != nil {
			return 0, "", fmt.Errorf("failed to read message content: %w", err)
		}
		readBytes += n
	}
	
	content := string(contentBuffer)
	return msgType, content, nil
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
		return numericResponse, fmt.Errorf("unexpected behavior in server: %w", err)
	}

	return numericResponse, err
}

// Close closes the underlying connection
func (ch *CommunicationHandler) Close() error {
	if ch.conn != nil {
		return ch.conn.Close()
	}
	return nil
}
