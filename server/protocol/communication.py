import logging
from .message import BetMessage

class CommunicationHandler:
    """Simple server-side communication handler"""
    
    def __init__(self, conn):
        self.conn = conn
    
    def receive_message(self):
        """Receive a length-prefixed batch message from the client with proper buffering"""
        try:
            # First read the total message size (8 characters for 8-digit length)
            size_bytes = self.conn.recv(8)
            if not size_bytes or len(size_bytes) < 8:
                raise ConnectionError("Connection closed by client or incomplete size")
            
            total_size = int(size_bytes.decode('utf-8'))
            logging.info(f"\n=== MESSAGE RECEPTION START ===\n")
            logging.info(f"Expected total message size: {total_size} bytes")
            logging.info(f"Size bytes (string): {size_bytes.decode('utf-8')}")
            
            # Read the complete message - keep reading until we get all bytes
            message_bytes = b''
            remaining_bytes = total_size
            
            logging.info(f"Need to read {remaining_bytes} additional bytes")
            
            while len(message_bytes) < remaining_bytes:
                chunk = self.conn.recv(remaining_bytes - len(message_bytes))
                if not chunk:
                    raise ConnectionError("Connection closed by client")
                message_bytes += chunk
                logging.info(f"Read chunk: {len(chunk)} bytes, total so far: {len(message_bytes)}/{remaining_bytes}")
                logging.info(f"Chunk content (repr): {repr(chunk)}")
            
            # Now we have the complete message, log it thoroughly
            logging.info(f"\n=== COMPLETE MESSAGE RECEIVED ===\n")
            logging.info(f"Total bytes received: {len(message_bytes) + 8} (including size bytes)")
            logging.info(f"Message bytes (repr): {repr(message_bytes)}")
            
            # Try to decode as UTF-8
            try:
                message_str = message_bytes.decode('utf-8')
                logging.info(f"Message as string: {repr(message_str)}")
                logging.info(f"Message length: {len(message_str)} characters")
            except UnicodeDecodeError as e:
                logging.error(f"Failed to decode message as UTF-8: {e}")
                logging.error(f"Raw bytes: {message_bytes}")
                raise Exception(f"Message encoding error: {e}")
            
            # Parse the batch message
            logging.info(f"=== STARTING MESSAGE PARSING ===")
            bets = self._parse_batch_message(message_str)
            
            logging.info(f"=== MESSAGE RECEPTION COMPLETE ===")
            logging.info(f"Successfully parsed {len(bets)} bets")
            
            return bets
            
        except Exception as e:
            logging.error(f"=== MESSAGE RECEPTION FAILED ===")
            logging.error(f"Error: {e}")
            raise Exception(f"Failed to receive message: {e}")
    
    def _parse_batch_message(self, message_str):
        """Parse batch bet message format: <bet size 1><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>..."""
        bets = []
        i = 0
        
        logging.info(f"=== PARSING BATCH MESSAGE ===")
        logging.info(f"Input message: {repr(message_str)}")
        logging.info(f"Message length: {len(message_str)} characters")
        
        while i < len(message_str):
            logging.info(f"--- Parsing at position {i} ---")
            logging.info(f"Remaining characters: {len(message_str) - i}")
            logging.info(f"Remaining substring: {repr(message_str[i:])}")
            
            # Need at least 2 characters for the bet size
            if i + 2 > len(message_str):
                logging.warning(f"Not enough characters remaining for bet size at position {i}")
                break
                
            # Read bet size (2 characters for 2-digit length)
            try:
                bet_size_chars = message_str[i:i+2]
                bet_size = int(bet_size_chars)
                logging.info(f"Position {i}: bet size characters '{bet_size_chars}' = {bet_size}")
            except ValueError:
                logging.warning(f"Invalid bet size characters '{message_str[i:i+2]}' at position {i}")
                break
                
            i += 2  # Skip 2 characters for the size
            logging.info(f"After reading size, position is now {i}")
            
            # Check if we have enough characters for the complete bet message
            if i + bet_size > len(message_str):
                logging.warning(f"Bet message truncated at position {i}, expected {bet_size} chars but only {len(message_str) - i} remaining")
                logging.warning(f"Remaining substring: {repr(message_str[i:])}")
                break
                
            # Extract the complete bet message
            bet_message = message_str[i:i+bet_size]
            logging.info(f"Position {i}: extracted bet message (length {bet_size}): {repr(bet_message)}")
            i += bet_size
            logging.info(f"After extracting bet message, position is now {i}")
            
            # Parse individual bet message
            try:
                logging.info(f"Attempting to parse bet message: {repr(bet_message)}")
                bet = BetMessage.bet_from_string(bet_message)
                bets.append(bet)
                logging.info(f"Successfully parsed bet: {bet}")
            except ValueError as e:
                logging.warning(f"Failed to parse bet message '{bet_message}': {e}")
                continue
        
        logging.info(f"=== PARSING COMPLETE ===")
        logging.info(f"Successfully parsed {len(bets)} bets")
        return bets
    
    def send_response(self, total_bets, stored_bets):
        """Send a simple response to the client"""
        try:
            # Add newline for simple text-based protocol
            if total_bets == stored_bets:
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {stored_bets}')
                full_response = str(stored_bets) + "\n"
            else:
                logging.info(f'action: apuesta_recibida | result: fail | cantidad: {stored_bets}')
                full_response = str(stored_bets-total_bets) + "\n"
            self.conn.send(full_response.encode('utf-8'))
        except Exception as e:
            raise Exception(f"Failed to send response: {e}")
    
    def close(self):
        """Close the connection"""
        if self.conn:
            self.conn.close()
