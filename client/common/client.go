package common

import (
	"fmt"
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"


	"github.com/op/go-logging"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	Bet           *domain.Bet
}

// Client Entity that encapsulates how
type Client struct {
	config  ClientConfig
	conn    net.Conn
	running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
// The signals are: SIGINT which is the Ctrl+C signal and SIGTERM which is the signal sent by the docker compose down command
// When one of these signals is received, the client stops sending messages and closes the connection to the server
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:  config,
		running: true,
	}

	signals := make(chan os.Signal, 1) //The 1 is the buffer size
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM) //Notify the client when the signals are received

	println("Client initialized")

	go func() {
		println("Waiting for signal")
		<-signals //Wait for the signal to be received
		println("Signal received, closing connection")
		client.running = false
		if client.conn != nil {
			client.conn.Close()
		}
		os.Exit(0)
	}()

	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err // Return error instead of continuing
	}
	c.conn = conn
	return nil
}

// submitBet submits a bet to the server using the simple string protocol
func (c *Client) submitBet(bet *domain.Bet) error {
	// Create communication handler
	commHandler := protocol.NewCommunicationHandler(c.conn)
	defer commHandler.Close()

	// Create bet message in the required format: <msg length><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>
	betMsg := protocol.NewBetMessage(
		c.config.ID, // agency ID is the client ID
		bet.Nombre,
		bet.Apellido,
		bet.Documento,
		bet.Nacimiento,
		bet.Numero,
	)

	// Send bet message
	if err := commHandler.SendMessage(betMsg.String()); err != nil {
		return fmt.Errorf("failed to send bet message: %w", err)
	}

	// Receive response
	response, err := commHandler.ReceiveMessage()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	// Parse response
	success, err := protocol.ParseResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if success {
		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %d", 
			bet.Documento, bet.Numero)
	} else {
		return fmt.Errorf("bet submission failed: %s", response)
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount && c.running; msgID++ {
		// Check if we should shutdown before creating connection
		if !c.running {
			log.Infof("action: client_shutdown | result: in_progress | client_id: %v", c.config.ID)
			break
		}

		// Create the connection to the server in every loop iteration
		if err := c.createClientSocket(); err != nil {
			log.Errorf("action: create_socket | result: fail | client_id: %v | error: %v", c.config.ID, err)
			// Wait before retrying
			time.Sleep(c.config.LoopPeriod)
			continue
		}

		// Check if connection was created successfully
		if c.conn == nil {
			log.Errorf("action: create_socket | result: fail | client_id: %v | error: connection is nil", c.config.ID)
			time.Sleep(c.config.LoopPeriod)
			continue
		}

		// Submit bet using the simple protocol
		if err := c.submitBet(c.config.Bet); err != nil {
			log.Errorf("action: submit_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			c.conn = nil
			time.Sleep(c.config.LoopPeriod)
			continue
		}

		// Close connection after successful bet submission
		c.conn.Close()
		c.conn = nil

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}
	
	if c.running {
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	} else {
		log.Infof("action: client_shutdown | result: success | client_id: %v", c.config.ID)
	}
}
