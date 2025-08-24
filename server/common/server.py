import socket
import logging
import os
import signal
import sys

from protocol.message import BetMessage
from protocol.communication import CommunicationHandler
from common.utils import store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        
        # Flag to control graceful shutdown
        self._running = True
        
        # Set up signal handlers ONCE during initialization
        self.__init_signals()
        
        # Set socket timeout to allow checking shutdown flag
        self._server_socket.settimeout(1.0)

    def __init_signals(self):
        """
        Initialize signal handlers for graceful shutdown
        This is called ONCE during server initialization
        """
        signal.signal(signal.SIGINT, self.__handle_signal)
        signal.signal(signal.SIGTERM, self.__handle_signal)

    def __handle_signal(self, signum, frame):
        """
        Handle shutdown signals gracefully
        This sets the shutdown flag instead of immediately exiting
        """
        logging.info(f'action: signal_received | result: success | signal: {signum}')
        self._running = False

    def __cleanup_resources(self):
        """
        Clean up all server resources gracefully
        """
        logging.info('action: cleanup_resources | result: in_progress')
        
        try:
            # Close server socket
            if hasattr(self, '_server_socket') and self._server_socket:
                self._server_socket.close()
                logging.info('action: cleanup_resources | result: success | resource: server_socket')
        except Exception as e:
            logging.error(f'action: cleanup_resources | result: fail | resource: server_socket | error: {e}')
        
        logging.info('action: cleanup_resources | result: success')

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        logging.info('action: server_start | result: success')
        
        try:
            while self._running: # This is the main loop of the server, waiting for new connections
                try:
                    client_sock = self.__accept_new_connection()
                    if client_sock:
                        self.__handle_client_connection(client_sock)
                except socket.timeout:
                    # Timeout allows checking shutdown flag - this is normal behavior, just continue
                    continue
                except Exception as e:
                    if self._running:
                        logging.error(f'action: accept_connection | result: fail | error: {e}')
                    break
        finally:
            logging.info('action: server_shutdown | result: in_progress')
            self.__cleanup_resources()
            logging.info('action: server_shutdown | result: success')

    def __handle_client_connection(self, client_sock):
        """
        Handle client connection using the bet protocol
        """
        try:
            # Create communication handler
            communication_handler = CommunicationHandler(client_sock)
            
            # Receive message from client
            message_str = communication_handler.receive_message()

            # Parse the bet message
            bet = BetMessage.bet_from_string(message_str)
            
            # Store the bet using the store_bets function
            store_bets([bet])
            
            # Log successful bet storage
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
            
            # Send success response to client (using bet number as ACK)
            communication_handler.send_response(str(bet.number))
            
        except Exception as e:
            addr = client_sock.getpeername() if client_sock else "unknown"
            logging.error(f"action: handle_client | result: fail | ip: {addr} | error: {e}")
            
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
