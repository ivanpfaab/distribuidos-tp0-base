import socket
import logging
import signal
import threading

from protocol.message import BetMessage
from protocol.communication import CommunicationHandler
from common.utils import store_bets, load_bets, has_won


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        
        # Flag to control graceful shutdown
        self._running = True
        
        # Thread-safe locks for shared resources
        self._completion_lock = threading.Lock()
        self._lottery_lock = threading.Lock()
        self._storage_lock = threading.Lock()
        
        # Track active client threads for graceful shutdown
        self._active_threads = set()
        self._threads_lock = threading.Lock()
        
        # Set up signal handlers ONCE during initialization
        self.__init_signals()
        
        # Set socket timeout to allow checking shutdown flag
        self._server_socket.settimeout(1.0)
        
        # Track completion notifications from agencies (thread-safe)
        self._completed_agencies = set()
        self._lottery_conducted = False
        self._winners_cache = {}  # Cache winners by agency
        self._active_connections = set()  # Track active client connections
        self._sending_bets_clients = {}  # Track client_id -> connection_socket mapping

    def __init_signals(self):
        """
        Initialize signal handlers for graceful shutdown
        This is called once during server initialization
        """
        signal.signal(signal.SIGINT, self.__handle_signal)
        signal.signal(signal.SIGTERM, self.__handle_signal)

    def __handle_signal(self, signum):
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
            # Wait for all active client threads to finish
            with self._threads_lock:
                active_threads = list(self._active_threads)
            
            if active_threads:
                logging.info(f'action: cleanup_resources | result: waiting_for_threads | active: {len(active_threads)}')
                
                # Wait for each thread to finish (with timeout)
                for thread in active_threads:
                    thread.join()  
                    if thread.is_alive():
                        logging.warning(f'action: cleanup_resources | result: thread_timeout | thread: {thread.name}')
                    else:
                        logging.info(f'action: cleanup_resources | result: thread_finished | thread: {thread.name}')
            
            # Close server socket
            if hasattr(self, '_server_socket') and self._server_socket:
                self._server_socket.close()
                logging.info('action: cleanup_resources | result: success | resource: server_socket')
                
        except Exception as e:
            logging.error(f'action: cleanup_resources | result: fail | error: {e}')
        
        logging.info('action: cleanup_resources | result: success')

    def __handle_notification(self, communication_handler, agency_id):
        """Handle completion notification from an agency (thread-safe)"""
        try:
            with self._completion_lock:
                self._completed_agencies.add(agency_id)
                logging.info(f'action: agency_completed | result: success | agency_id: {agency_id}')
            
            # Don't conduct lottery here - wait for client to disconnect
            # Send acknowledgment
            communication_handler.send_notification_response(True)
            
        except Exception as e:
            logging.error(f'action: handle_notification | result: fail | agency_id: {agency_id} | error: {e}')
            communication_handler.send_notification_response(False)

    def __conduct_lottery(self):
        """Conduct the lottery and find winners for each agency (thread-safe)"""
        # Use lock to ensure only one thread conducts the lottery
        with self._lottery_lock:
            # Double-check if lottery was already conducted
            if self._lottery_conducted:
                return
                
            try:
                logging.info('action: sorteo | result: success')
                
                # Load all bets and find winners
                all_bets = list(load_bets())
                winners_by_agency = {}
                
                for bet in all_bets:
                    if has_won(bet):
                        if bet.agency not in winners_by_agency:
                            winners_by_agency[bet.agency] = []
                        winners_by_agency[bet.agency].append(bet.document)
                
                # Cache winners by agency
                self._winners_cache = winners_by_agency
                self._lottery_conducted = True
                
                logging.info(f'action: lottery_winners_found | result: success | total_winners: {sum(len(winners) for winners in winners_by_agency.values())}')
                
            except Exception as e:
                logging.error(f'action: conduct_lottery | result: fail | error: {e}')

    def __handle_winner_query(self, communication_handler, agency_id):
        """Handle winner query from an agency (thread-safe)"""
        try:
            # Check if all active connections have completed sending bets
            with self._completion_lock:
                completed_count = len(self._completed_agencies)
                active_count = len(self._active_connections)
            
            if completed_count < active_count:
                # Not all active clients have completed yet
                logging.info(f'action: winner_query | result: waiting | agency_id: {agency_id} | completed: {completed_count} | active: {active_count}')
                communication_handler.send_waiting_response()
                return
            
            # All active clients have completed, proceed with lottery if not done yet
            if not self._lottery_conducted:
                self.__conduct_lottery()
            
            # Send winners to this agency
            with self._lottery_lock:
                if agency_id in self._winners_cache:
                    winners = self._winners_cache[agency_id]
                else:
                    winners = []
            
            communication_handler.send_winner_list(winners)
            logging.info(f'action: winner_query | result: success | agency_id: {agency_id} | winners_count: {len(winners)}')
            
        except Exception as e:
            logging.error(f'action: handle_winner_query | result: fail | agency_id: {agency_id} | error: {e}')
            communication_handler.send_waiting_response()

    def run(self):
        """
        Main server loop that accepts connections and creates a thread for each

        Server that accept new connections and establishes a
        communication with a client using parallel processing.
        """
        logging.info('action: server_start | result: success')
        
        try:
            while self._running: # This is the main loop of the server, waiting for new connections
                try:
                    client_sock = self.__accept_new_connection()
                    if client_sock:
                        # Create a new thread for each client connection
                        client_thread = threading.Thread(
                            target=self.__handle_client_connection, 
                            args=(client_sock,),
                            daemon=False,  # Non-daemon so we can wait for it
                            name=f"ClientHandler-{client_sock.getpeername()[0]}"
                        )
                        
                        # Track the thread for graceful shutdown
                        with self._threads_lock:
                            self._active_threads.add(client_thread)
                        
                        client_thread.start()
                        
                except socket.timeout:
                    # Timeout allows checking shutdown flag - this is normal behavior, just continue
                    continue
                except Exception as e:
                    # Log error but continue accepting connections
                    logging.error(f"action: accept_connection | result: fail | error: {e}")
                    continue
                    
        finally:
            logging.info('action: server_shutdown | result: in_progress')
            self.__cleanup_resources()
            logging.info('action: server_shutdown | result: success')

    def __handle_client_connection(self, client_sock):
        """
        Handle client connection using the bet protocol (runs in separate thread)
        """
        try:
            # Create communication handler
            communication_handler = CommunicationHandler(client_sock)
            
            # Track this connection (thread-safe)
            with threading.Lock():
                self._active_connections.add(client_sock)
            
            # Keep connection open to handle multiple messages from the same client
            while self._running:  # Check shutdown flag
                try:
                    # Receive message from client - this returns a dict with message type and data
                    message_data = communication_handler.receive_message()
                    
                    if message_data["type"] == "notification":
                        # Handle completion notification
                        self.__handle_notification(communication_handler, message_data["agency_id"])
                        
                    elif message_data["type"] == "winner_query":
                        # Handle winner query
                        self.__handle_winner_query(communication_handler, message_data["agency_id"])
                        
                    elif message_data["type"] == "batch_bets":
                        # Handle batch bet submission (thread-safe storage)
                        bets = message_data["bets"]
                        bets_stored = 0
                        
                        # Process each bet object with thread-safe storage
                        with self._storage_lock:
                            for bet in bets:
                                if bet:
                                    # Store the bet using the store_bets function
                                    store_bets([bet])
                                    bets_stored += 1
                        
                        # Send success response to client
                        # first number is the total number of bets
                        # second number is the number of bets stored
                        communication_handler.send_response(len(bets), bets_stored)
                    
                except ConnectionResetError as e:
                    # Client has finished sending all data and closed connection
                    # This is normal behavior, not an error
                    logging.info(f"action: client_disconnected | result: success | detail: client finished sending data")
                    break
                except Exception as e:
                    # Client disconnected - this is normal when they finish sending data
                    logging.info(f"action: client_disconnected | result: success | detail: client finished sending data")
                    break
                    
        except Exception as e:
            addr = client_sock.getpeername() if client_sock else "unknown"
            logging.error(f"action: handle_client | result: fail | ip: {addr} | error: {e}")
            
        finally:
            # Remove from active connections and close (thread-safe)
            with threading.Lock():
                self._active_connections.discard(client_sock)
                
                # Find and remove this client from connected_clients
                client_id_to_remove = None
                for client_id, sock in self._sending_bets_clients.items():
                    if sock == client_sock:
                        client_id_to_remove = client_id
                        break
                
                if client_id_to_remove:
                    del self._sending_bets_clients[client_id_to_remove]
                    logging.info(f'action: client_sent_all_bets | result: success | client_id: {client_id_to_remove}')
                
                # Check if all clients have been processed (disconnected)
                if len(self._sending_bets_clients) == 0:
                    self.__conduct_lottery()
            
            # Remove this thread from active threads
            with self._threads_lock:
                self._active_threads.discard(threading.current_thread())
            
            communication_handler.close()

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
        
        # Assign a unique client ID and track this connection (thread-safe)
        with threading.Lock():
            client_id = addr[0]
            self._sending_bets_clients[client_id] = c
            logging.info(f'action: new_client_connected | result: success | client_number: {client_id}')
            
            # Track this connection
            self._active_connections.add(c)
        
        return c
