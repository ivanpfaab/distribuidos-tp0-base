package common

import (
	"fmt"
	"io"
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strconv"


	"github.com/op/go-logging"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            	string
	ServerAddress 	string
	LoopAmount    	int
	LoopPeriod    	time.Duration
	FilePath	  	string
	MaxBatchAmount 	int
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

	go func() {
		<-signals //Wait for the signal to be received
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
func (c *Client) submitBets(bets []*domain.Bet) error {
	// Create communication handler
	commHandler := protocol.NewCommunicationHandler(c.conn)

	betMsg := protocol.NewBatchBetMessage(bets)

	// Send bet message
	if err := commHandler.SendMessage(betMsg.FormatBatch()); err != nil {
		return fmt.Errorf("failed to send bet message: %w", err)
	}

	// Receive response
	ack, err := commHandler.ReceiveMessage()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	if ack == len(bets) {
		log.Infof("action: batch_sent | result: success | cantidad: %d", len(bets))
	} else {
		log.Errorf("action: batch_sent | result: fail | cantidad: %d", ack)
		return fmt.Errorf("server returned non-positive response: %d", ack)
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {

	file, err := os.Open(c.config.FilePath)
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	defer file.Close()
	chunkReader := NewCSVChunkReader(file)

	loopCount := 0
	for ; loopCount < c.config.LoopAmount; loopCount++ {
		// Create the connection to the server
		if err := c.createClientSocket(); err != nil {
			log.Errorf("action: create_socket | result: fail | client_id: %v | error: %v", c.config.ID, err)
			// Wait before retrying
			time.Sleep(c.config.LoopPeriod)
			continue
		}

		// Check if connection was created successfully
		if c.conn == nil {
			log.Errorf("action: create_socket | result: fail | client_id: %v | error: connection is nil", c.config.ID, err)
			time.Sleep(c.config.LoopPeriod)
			continue
		}

		break
	}

	if loopCount >= c.config.LoopAmount {
		log.Errorf("action: create_socket | result: fail | client_id: %v | error: max retries reached", c.config.ID)
		return
	}

	// Process data in batches 
	var currentBatch [][]string
	
	for chunkReader.HasMore() && c.running {
		// Read a chunk and add to current batch
		chunk, err := chunkReader.ReadChunk(c.config.MaxBatchAmount) 
		if err != nil && err != io.EOF {
			log.Errorf("action: read_chunk | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		
		if err == io.EOF {
			break
		}
		
		// Add chunk to current batch
		currentBatch = append(currentBatch, chunk...)
		
		// If we've reached MaxBatchAmount or there's no more data, process the batch
		if len(currentBatch) >= c.config.MaxBatchAmount || !chunkReader.HasMore() {
			// Convert client ID to integer for agency ID
			agencyID, err := strconv.Atoi(c.config.ID)
			if err != nil {
				log.Errorf("action: parse_client_id | result: fail | client_id: %v | error: %v", c.config.ID, err)
				return
			}
			
			bets, err := domain.BetsFromChunk(currentBatch, agencyID)
			if err != nil {
				log.Errorf("action: parse_chunk | result: fail | client_id: %v | error: %v", c.config.ID, err)
				return
			}

			// Submit bets using batch request
			if err := c.submitBets(bets); err != nil {
				log.Errorf("action: submit_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
				time.Sleep(c.config.LoopPeriod)
				continue
			}

			// Clear current batch and wait before next batch
			currentBatch = currentBatch[:0]
			time.Sleep(c.config.LoopPeriod)
		}
	}

	// Close connection only after all bets have been processed
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.running = false

	if c.running {
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	} else {
		log.Infof("action: client_shutdown | result: success | client_id: %v", c.config.ID)
	}
}
