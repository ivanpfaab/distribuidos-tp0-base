class CommunicationHandler:
    """Simple server-side communication handler"""
    
    def __init__(self, conn):
        self.conn = conn
    
    def receive_message(self):
        """Receive a length-prefixed message from the client"""
        try:
            # First read the length byte
            length_bytes = self.conn.recv(1)
            if not length_bytes:
                raise ConnectionError("Connection closed by client")
            
            length = length_bytes[0]
            
            # Then read exactly that many bytes for the message
            bets = []
            message_bytes = b''
            while len(message_bytes) < length:
                sub_length = self.conn.recv(1)
                if not sub_length:
                    continue
                message_bytes += sub_length
                sub_length = sub_length[0]
                chunk = self.conn.recv(length - len(sub_length))
                
                if not chunk:
                    raise ConnectionError("Connection closed by client")
                message_bytes += chunk
                bets.append(chunk.decode('utf-8'))
            
            # return array of bets in string format
            return bets
            
        except Exception as e:
            raise Exception(f"Failed to receive message: {e}")
    
    def send_response(self, total_bets, stored_bets):
        """Send a simple response to the client"""
        try:
            # Add newline for simple text-based protocol
            if total_bets == stored_bets:
                logging.info(f'action: apuesta_recibida | result: success | cantidad: ${stored_bets}')
                full_response = stored_bets + "\n"
            else:
                logging.info(f'action: apuesta_recibida | result: fail | cantidad: ${stored_bets}')
                full_response = str(stored_bets-total_bets) + "\n"
            self.conn.send(full_response.encode('utf-8'))
        except Exception as e:
            raise Exception(f"Failed to send response: {e}")
    
    def close(self):
        """Close the connection"""
        if self.conn:
            self.conn.close()
