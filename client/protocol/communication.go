package protocol

import (
	"fmt"
	"net"
	"strings"
)

// CommunicationHandler handles communication with the new protocol format
type CommunicationHandler struct {
	conn net.Conn
}

// NewCommunicationHandler creates a new communication handler
func NewCommunicationHandler(conn net.Conn) *CommunicationHandler {
	return &CommunicationHandler{
		conn: conn,
	}
}

// SendMessage sends a message with the new protocol format: <type><size><content>
func (ch *CommunicationHandler) SendMessage(msgType byte, content string) error {	
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

// SendBatchBets sends a batch of bets to the server
func (ch *CommunicationHandler) SendBatchBets(content string) error {
	return ch.SendMessage(MessageTypeBatchBets, content)
}

// SendNotification sends a completion notification to the server
func (ch *CommunicationHandler) SendNotification(agencyID int) error {
	notification := fmt.Sprintf("%d", agencyID)
	return ch.SendMessage(MessageTypeNotification, notification)
}

// SendWinnerQuery sends a winner query to the server
func (ch *CommunicationHandler) SendWinnerQuery(agencyID int) error {
	query := fmt.Sprintf("%d", agencyID)
	return ch.SendMessage(MessageTypeWinnerQuery, query)
}

// receiveMessageWithProtocol receives a message using the new protocol format: <type><size><content>
func (ch *CommunicationHandler) receiveMessageWithProtocol() (byte, string, error) {
	// First read the message type (1 byte)
	typeBuffer := make([]byte, 1)
	_, err := ch.conn.Read(typeBuffer)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read message type: %w", err)
	}
	msgType := typeBuffer[0]
	
	// Then read the message size (8 characters for 8-digit length)
	sizeBuffer := make([]byte, 8)
	_, err = ch.conn.Read(sizeBuffer)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read message size: %w", err)
	}
	
	var totalSize int
	_, err = fmt.Sscanf(string(sizeBuffer), "%d", &totalSize)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse message size: %w", err)
	}
	
	// Read the complete message content
	contentBuffer := make([]byte, totalSize)
	readBytes := 0
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

// ReceiveMessage receives a response message from the server (for batch bet responses)
func (ch *CommunicationHandler) ReceiveMessage() (int, error) {
	msgType, content, err := ch.receiveMessageWithProtocol()
	if err != nil {
		return -1, fmt.Errorf("failed to receive message: %w", err)
	}
	
	// Expect response type 'R' for batch bet responses
	if msgType != 'R' {
		return -1, fmt.Errorf("unexpected message type: %c, expected 'R'", msgType)
	}
	
	parsedResponse, err := ParseResponse(content)
	return parsedResponse, err
}

// ReceiveNotificationResponse receives a notification acknowledgment from the server
func (ch *CommunicationHandler) ReceiveNotificationResponse() (bool, error) {
	msgType, content, err := ch.receiveMessageWithProtocol()
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
func (ch *CommunicationHandler) ReceiveWinnerList() ([]string, error) {
	msgType, content, err := ch.receiveMessageWithProtocol()
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

// Close closes the underlying connection
func (ch *CommunicationHandler) Close() error {
	if ch.conn != nil {
		return ch.conn.Close()
	}
	return nil
}
