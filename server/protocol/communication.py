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
            message_bytes = b''
            while len(message_bytes) < length:
                chunk = self.conn.recv(length - len(message_bytes))
                if not chunk:
                    raise ConnectionError("Connection closed by client")
                message_bytes += chunk
            
            # Convert bytes to string
            return message_bytes.decode('utf-8')
            
        except Exception as e:
            raise Exception(f"Failed to receive message: {e}")
    
    def send_response(self, response):
        """Send a simple response to the client"""
        try:
            # Add newline for simple text-based protocol
            full_response = response + "\n"
            self.conn.send(full_response.encode('utf-8'))
        except Exception as e:
            raise Exception(f"Failed to send response: {e}")
    
    def close(self):
        """Close the connection"""
        if self.conn:
            self.conn.close()
