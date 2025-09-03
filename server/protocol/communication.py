import logging
from .message import BetMessage

# Message type constants for server responses
MESSAGE_TYPE_RESPONSE = 'R'        # Response to batch bet submission
MESSAGE_TYPE_ACKNOWLEDGMENT = 'A'  # Acknowledgment for notifications
MESSAGE_TYPE_WINNERS = 'W'         # Winner list response
MESSAGE_TYPE_WAITING = 'T'         # Waiting for other clients to complete

class CommunicationHandler:
    """Server-side communication handler with new protocol format"""
    
    def __init__(self, conn):
        self.conn = conn
    
    def receive_message(self):
        """Receive a message with new protocol format: <type><size><content>"""
        try:
            # First read the message type (1 byte)
            type_byte = self.conn.recv(1)
            if not type_byte:
                raise ConnectionResetError("Connection closed by client or incomplete type")
            
            message_type = type_byte.decode('utf-8')
            logging.debug(f"Received message type: {message_type}")
            
            # Then read the message size (8 characters for 8-digit length)
            size_bytes = self.conn.recv(8)
            if not size_bytes or len(size_bytes) < 8:
                raise ConnectionResetError("Connection closed by client or incomplete size")
            
            total_size = int(size_bytes.decode('utf-8'))
            logging.debug(f"Message content size: {total_size} bytes")
            
            # Read the complete message content
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
            
            # Parse message based on type
            if message_type == 'B':
                return self._parse_batch_message(message_str)
            elif message_type == 'N':
                return self._parse_notification_message(message_str)
            elif message_type == 'W':
                return self._parse_winner_query_message(message_str)
            else:
                raise ValueError(f"Unknown message type: {message_type}")
            
        except Exception as e:
            raise Exception(f"Failed to receive message: {e}")
    
    def _parse_notification_message(self, message_str):
        """Parse notification message format: agency_id"""
        try:
            agency_id = int(message_str)
            return {"type": "notification", "agency_id": agency_id}
        except Exception as e:
            raise ValueError(f"Failed to parse notification message: {e}")
    
    def _parse_winner_query_message(self, message_str):
        """Parse winner query message format: agency_id"""
        try:
            agency_id = int(message_str)
            return {"type": "winner_query", "agency_id": agency_id}
        except Exception as e:
            raise ValueError(f"Failed to parse winner query message: {e}")
    
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
        return {"type": "batch_bets", "bets": bets}
    
    def _send_message(self, msg_type, content):
        """Send a message using the protocol format: <type><size><content>"""
        try:
            size_str = f"{len(content):08d}"
            full_message = msg_type + size_str + content
            message_bytes = full_message.encode('utf-8')
            
            
            written_bytes = 0
            while written_bytes < len(message_bytes):
                sent = self.conn.send(message_bytes[written_bytes:])
                written_bytes += sent
                
        except Exception as e:
            raise Exception(f"Failed to send message: {e}")
    
    def send_batch_response(self, total_bets, stored_bets):
        """Send a response to batch bet submission"""
        try:
            # Check the status depending on how many bets were stored
            execution_status = "success" if total_bets == stored_bets else "fail"
            logging.info(f'action: apuesta_recibida | result: {execution_status} | cantidad: {stored_bets}')
            
            # Send response using protocol format: 'R' for response
            response_content = str(stored_bets)
            self._send_message(MESSAGE_TYPE_RESPONSE, response_content)
            
        except Exception as e:
            raise Exception(f"Failed to send response: {e}")
    
    def send_notification_response(self, success):
        """Send response to notification message"""
        try:
            # Send response using protocol format: 'A' for acknowledgment
            response_content = "1" if success else "0"
            self._send_message(MESSAGE_TYPE_ACKNOWLEDGMENT, response_content)
            
        except Exception as e:
            raise Exception(f"Failed to send notification response: {e}")
    
    def send_waiting_response(self):
        """Send response indicating server is waiting for other clients"""
        try:
            # Send waiting response using protocol format: 'T' for waiting
            self._send_message(MESSAGE_TYPE_WAITING, "waiting")
            
        except Exception as e:
            raise Exception(f"Failed to send waiting response: {e}")
    
    def send_winner_list(self, winners):
        """Send winner list to client"""
        try:
            # Send winner list using protocol format: 'W' for winners
            if not winners:
                winner_content = ""
            else:
                winner_content = ",".join(winners)
            
            self._send_message(MESSAGE_TYPE_WINNERS, winner_content)
            
            logging.info(f'action: winners_sent | result: success | winners_count: {len(winners)}')
            
        except Exception as e:
            raise Exception(f"Failed to send winner list: {e}")
    
    def close(self):
        """Close the connection"""
        if self.conn:
            self.conn.close()
