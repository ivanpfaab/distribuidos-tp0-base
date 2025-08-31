# Ejercicio 7: Notificación de Finalización y Consulta de Ganadores

## Problema a Resolver

### Objetivo
Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas, permitiendo que el servidor proceda con el sorteo. Los clientes deben consultar inmediatamente después la lista de ganadores correspondientes a su agencia.

### Requisitos Específicos
- **Cliente**: Notificar al servidor al finalizar el envío de todas las apuestas
- **Cliente**: Consultar lista de ganadores del sorteo correspondientes a su agencia
- **Cliente**: Imprimir log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`
- **Servidor**: Esperar notificación de las 5 agencias antes de realizar el sorteo
- **Servidor**: Imprimir log: `action: sorteo | result: success`
- **Servidor**: Verificar apuestas con `load_bets(...)` y `has_won(...)` (funciones provistas)
- **Servidor**: Retornar solo los DNIs ganadores de la agencia consultante
- **Restricción**: No realizar broadcast de todos los ganadores hacia todas las agencias
- **Restricción**: No responder consultas de ganadores con información parcial antes del sorteo

## Diseño de la Solución

### Arquitectura de la Solución

La solución implementa un **protocolo de comunicación robusto** que:

1. **Cliente envía todas las apuestas** en batches
2. **Cliente notifica finalización** al servidor
3. **Cliente se desconecta** para permitir que otros clientes procesen
4. **Cliente se reconecta** para consultar ganadores
5. **Servidor espera** notificaciones de todas las agencias
6. **Servidor realiza sorteo** cuando todas han completado
7. **Servidor responde consultas** con ganadores específicos por agencia

### Flujo de Comunicación

```
Cliente 1: [Bets] → [Notification] → [Disconnect] → [Reconnect] → [Winner Query] → [Winners + Documents]
Cliente 2: [Bets] → [Notification] → [Disconnect] → [Reconnect] → [Winner Query] → [Winners + Documents]
...
Cliente 5: [Bets] → [Notification] → [Disconnect] → [Reconnect] → [Winner Query] → [Winners + Documents]
Servidor: [Wait for all] → [Conduct Lottery] → [Respond Queries]
```

## Implementación Técnica

### 1. Protocolo de Comunicación Mejorado

#### Estructura de Mensajes
Se implementó un protocolo robusto con prefijo de tipo y tamaño:

```
<message_type_byte><msg_size_8_bytes><message_content>
```

#### Tipos de Mensaje
- **'B' (Batch Bets)**: Envío de apuestas en lote
- **'N' (Notification)**: Notificación de finalización
- **'W' (Winner Query)**: Consulta de ganadores
- **'R' (Response)**: Respuesta a batch de apuestas
- **'A' (Acknowledgment)**: Confirmación de notificación
- **'W' (Winners)**: Lista de ganadores
- **'T' (Waiting)**: Servidor esperando otros clientes
- **'D' (Documents)**: Lista de documentos de la agencia

### 2. Modificaciones del Cliente

#### Estrategia de Desconexión/Reconexión
```go
// Después de enviar todas las apuestas y notificar
if c.conn != nil {
    c.conn.Close()
    c.conn = nil
}

// Esperar para dar tiempo a otros clientes
time.Sleep(c.config.LoopPeriod * 2)

// Reconectar para consultar ganadores
if c.running {
    if err := c.queryWinnersWithReconnect(); err != nil {
        log.Errorf("action: query_winners | result: fail | client_id: %v | error: %v", c.config.ID, err)
    }
}
```


### 3. Modificaciones del Servidor

#### Tracking Dinámico de Clientes
```python
class Server:
    def __init__(self, port, listen_backlog):
        # ... socket init ...
        self._completed_agencies = set()  # Agencias que notificaron finalización
        self._unique_agencies_ever_connected = set()  # Todas las agencias únicas que enviaron apuestas
        self._lottery_conducted = False  # Flag de sorteo realizado
        self._winners_cache = {}  # Cache de ganadores por agencia
        self._active_connections = set()  # Conexiones activas actualmente
```


## Caracteristicas de la Solución

### 1. **Cambio de sintaxis en protocolo**
- Prefijo de tipo y tamaño para interpretación clara
- Separación clara entre diferentes tipos de mensaje

### 2. **Tracking Dinámico de Clientes**
- No asume un número fijo de agencias
- Aprende dinámicamente cuántas agencias únicas existen

### 3. **Estrategia de Desconexión/Reconexión**
- Permite que múltiples clientes procesen secuencialmente
- Evita bloqueos en el servidor de procesamiento secuencial
- Da tiempo para que otros clientes suban sus datos


## Instrucciones de Ejecución

### 1. Construir y Levantar el Sistema
```bash
make docker-compose-up
```

### 2. Ver Logs en Tiempo Real
```bash
make docker-compose-logs
```

### 3. Detener el Sistema
```bash
make docker-compose-down
```

### 4. Reconstruir Imágenes (si es necesario)
```bash
make docker-image
```

## Ejemplo de Logs Esperados

### Cliente
```
client1  | action: batch_sent | result: success | cantidad: 4
client1  | action: notification_sent | result: success | client_id: 1
client1  | action: waiting_for_other_clients | result: success | client_id: 1
client1  | action: consulta_ganadores | result: success | cant_ganadores: 0
client1  | action: documentos_recibidos | result: success | client_id: 1 | cant_documentos: 10
```

### Servidor
```
server   | action: agency_completed | result: success | agency_id: 1
server   | action: sorteo | result: success
server   | action: lottery_winners_found | result: success | total_winners: 3
server   | action: winner_query | result: success | agency_id: 1 | winners_count: 0
server   | action: documents_sent | result: success | agency_id: 1 | document_count: 10
```