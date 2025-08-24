from common.utils import Bet


class BetMessage:
    """Simple bet message parser for the server"""
    
    def __init__(self, agency_id, nombre, apellido, documento, nacimiento, numero):
        self.agency_id = agency_id
        self.nombre = nombre
        self.apellido = apellido
        self.documento = documento
        self.nacimiento = nacimiento
        self.numero = numero
    
    @classmethod
    def bet_from_string(cls, message_str) -> Bet:
        """
        Parse bet message from string format: agency_id|nombre|apellido|documento|nacimiento|numero

        """
        try:
            # Split the message by | to get each field
            fields = message_str.split('|')
            
            if len(fields) != 6:
                raise ValueError(f"Expected 6 fields, got {len(fields)}")
            
            agency_id_str, nombre, apellido, documento, nacimiento, numero_str = fields
            
            # Convert numero to integer and agency_id to integer
            try:
                numero = int(numero_str)
                agency_id = int(agency_id_str)
            except ValueError:
                raise ValueError(f"Invalid numero format: {numero_str}")
            
            bet = Bet(agency_id, nombre, apellido, documento, nacimiento, numero)
            return bet
            
        except Exception as e:
            raise ValueError(f"Failed to parse bet message: {e}")
    
    def __str__(self):
        """String representation of the bet message"""
        return f"Bet{{agency: {self.agency_id}, dni: {self.documento}, numero: {self.numero}, name: {self.nombre} {self.apellido}, birth: {self.nacimiento}}}"
