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
                raise ConnectionResetError("Connection closed by client or incomplete size")
            
            total_size = int(size_bytes.decode('utf-8'))
            logging.info(f"Total message size: {total_size} bytes")
            
            # Read the complete message - keep reading until we get all bytes
            message_bytes = b''
            remaining_bytes = total_size
            
            while len(message_bytes) < remaining_bytes:

                chunk = self.conn.recv(remaining_bytes - len(message_bytes))
                if not chunk:
                    raise ConnectionResetError("Connection closed by client")
                message_bytes += chunk
            
            # Try to decode as UTF-8
            try:
                message_str = message_bytes.decode('utf-8')
            except UnicodeDecodeError as e:
                raise Exception(f"Message encoding error: {e}")
            
            bets = self._parse_batch_message(message_str)
            
            return bets
            
        except Exception as e:
            raise Exception(f"Failed to receive message: {e}")
    
    def _parse_batch_message(self, message_str):
        """Parse batch bet message format: <agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>&&<agency id>..."""
        # Remove the last element of the list since it's an extra from the split
        bets_str = message_str.split('&&')[:-1]
        bets = []
        
        for bet_str in bets_str:
            # Parse individual bet message
            try:
                bet = BetMessage.bet_from_string(bet_str)
                bets.append(bet)
                
            except ValueError as e:
                logging.warning(f"Failed to parse bet message '{bet_str}': {e}")
                continue
        return bets
    
    def send_response(self, total_bets, stored_bets):
        """Send a simple response to the client"""
        try:
            # Check the status depending on how many bets were stored
            execution_status = "success" if total_bets == stored_bets else "fail"
            logging.info(f'action: apuesta_recibida | result: {execution_status} | cantidad: {stored_bets}')
            
            full_response = str(stored_bets) + "\n"
            message = full_response.encode('utf-8')
            
            # Send all data, handling partial sends
            total_sent = 0
            while total_sent < len(message):
                sent = self.conn.send(message[total_sent:])
                total_sent += sent
                
        except Exception as e:
            raise Exception(f"Failed to send response: {e}")
    
    def close(self):
        """Close the connection"""
        if self.conn:
            self.conn.close()
