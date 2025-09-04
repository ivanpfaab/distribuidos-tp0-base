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

**Componentes:**
- `message_type_byte`: 1 byte que identifica el tipo de mensaje
- `msg_size_8_bytes`: 8 bytes que indican el tamaño del contenido (formato: "00000000"-"99999999")
- `message_content`: Contenido real del mensaje (máximo 8KB)

#### Tipos de Mensaje del Cliente al Servidor

| Tipo | Código | Descripción | Contenido |
|------|--------|-------------|-----------|
| **Batch Bets** | `'B'` | Envío de apuestas en lote | `<agency_id>|<nombre>|<apellido>|<documento>|<nacimiento>|<numero>&&...` |
| **Notification** | `'N'` | Notificación de finalización | `<agency_id>` |
| **Winner Query** | `'W'` | Consulta de ganadores | `<agency_id>` |

#### Tipos de Mensaje del Servidor al Cliente

| Tipo | Código | Descripción | Contenido |
|------|--------|-------------|-----------|
| **Response** | `'R'` | Respuesta a batch de apuestas | `<cantidad_almacenada>` |
| **Acknowledgment** | `'A'` | Confirmación de notificación | `"1"` (éxito) o `"0"` (fallo) |
| **Winners** | `'W'` | Lista de ganadores | `<dni1>,<dni2>,<dni3>...` (vacío si no hay ganadores) |
| **Waiting** | `'T'` | Servidor esperando otros clientes | `"waiting"` |

### 2. Flujo de Interacción Detallado

#### Fase 1: Envío de Apuestas
```
Cliente → Servidor: 'B' + tamaño + contenido_batch
Servidor → Cliente: 'R' + cantidad_almacenada
```

#### Fase 2: Notificación de Finalización
```
Cliente → Servidor: 'N' + agency_id
Servidor → Cliente: 'A' + "1" (confirmación)
Cliente: Desconecta
```

#### Fase 3: Consulta de Ganadores
```
Cliente: Reconecta
Cliente → Servidor: 'W' + agency_id
Servidor → Cliente: 'W' + lista_ganadores O 'T' + "waiting"
Cliente: Desconecta
```

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
        self._active_connections = set()  # Conexiones activas actualmente
        self._lottery = Lottery()  # Sistema de lotería
```

#### Manejo de Estados del Servidor
```python
def __handle_notification(self, communication_handler, agency_id):
    """Maneja notificación de finalización de una agencia"""
    self._completed_agencies.add(agency_id)
    logging.info(f'action: agency_completed | result: success | agency_id: {agency_id}')
    communication_handler.send_notification_response(True)

def __handle_winner_query(self, communication_handler, agency_id):
    """Maneja consulta de ganadores de una agencia"""
    # Verificar si todas las agencias activas han completado
    if len(self._completed_agencies) < len(self._active_connections):
        communication_handler.send_waiting_response()
        return
    
    # Realizar sorteo si no se ha hecho
    if not self._lottery.is_lottery_conducted():
        self._lottery.conduct_lottery()
    
    # Enviar ganadores específicos de la agencia
    winners = self._lottery.get_winners_for_agency(agency_id)
    communication_handler.send_winner_list(winners)
```

## Características de la Solución

### 1. **Protocolo Robusto con Tipos de Mensaje**
- **Prefijo de tipo**: Identificación clara del tipo de mensaje
- **Prefijo de tamaño**: Manejo seguro de mensajes de cualquier longitud
- **Separación de responsabilidades**: Cada tipo de mensaje tiene un propósito específico
- **Manejo de errores**: Validación de tipos y tamaños de mensaje

### 2. **Tracking Dinámico de Clientes**
- **No asume número fijo**: El servidor aprende dinámicamente cuántas agencias existen
- **Conexiones activas**: Rastrea conexiones actualmente abiertas
- **Agencias completadas**: Mantiene registro de qué agencias han notificado finalización
- **Sincronización**: Solo procede con el sorteo cuando todas las agencias activas han completado

### 3. **Estrategia de Desconexión/Reconexión**
- **Procesamiento secuencial**: Permite que múltiples clientes procesen sin conflictos
- **Liberación de recursos**: Desconecta después de enviar apuestas para liberar conexiones
- **Reconexión inteligente**: Se reconecta solo para consultar ganadores
- **Retry con backoff**: Implementa reintentos con delay exponencial para consultas de ganadores

### 4. **Manejo de Estados del Servidor**
- **Estado de lotería**: Controla si el sorteo ya se realizó
- **Cache de ganadores**: Almacena ganadores agrupados por agencia
- **Respuestas condicionales**: Responde "waiting" si no todas las agencias han completado
- **Integridad de datos**: Solo permite consultas después de que todas las agencias hayan terminado

### 5. **Robustez de Red**
- **Manejo de partial reads/writes**: Implementa loops para garantizar transmisión completa
- **Timeouts y reconexión**: Maneja desconexiones inesperadas
- **Validación de mensajes**: Verifica tipos y tamaños antes de procesar
- **Logging detallado**: Registra todas las operaciones para debugging

### 6. **Seguridad y Privacidad**
- **Ganadores por agencia**: Cada agencia solo recibe sus propios ganadores
- **No broadcast**: Evita enviar información de otras agencias
- **Validación de agencia**: Verifica que las consultas provengan de agencias válidas


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
client1  | action: winner_query | result: waiting | client_id: 1 | attempt: 1
client1  | action: consulta_ganadores | result: success | cant_ganadores: 2
client1  | action: client_shutdown | result: success | client_id: 1
```

### Servidor
```
server   | action: new_client_connected | result: success | client_number: 172.18.0.3
server   | action: apuesta_recibida | result: success | cantidad: 4
server   | action: agency_completed | result: success | agency_id: 1
server   | action: client_sent_all_bets | result: success | client_id: 172.18.0.3
server   | action: sorteo | result: success
server   | action: winner_query | result: success | agency_id: 1 | winners_count: 2
server   | action: winners_sent | result: success | winners_count: 2
```