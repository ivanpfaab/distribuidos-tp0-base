# Ejercicio 5: Sistema de Quiniela con Protocolo de Comunicación

## Problema a Resolver

### Antes de la Solución
- El cliente y servidor usaban un protocolo simple de "echo" basado en strings
- No había separación entre la lógica de negocio y la capa de comunicación
- Los mensajes se enviaban sin estructura definida
- No había manejo de datos estructurados (apuestas de quiniela)
- **Resultado**: Sistema básico sin capacidad para manejar apuestas reales

### Después de la Solución
- Implementación de un protocolo estructurado para apuestas de quiniela
- Separación clara entre dominio de negocio y comunicación
- Manejo robusto de sockets para evitar short read/write
- Sistema completo de envío y confirmación de apuestas
- **Resultado**: Sistema funcional para agencias de quiniela

## Diseño de la Solución

### Arquitectura del Sistema

```
┌─────────────────┐    Protocolo de Comunicación     ┌─────────────────┐
│   Cliente       │ ←──────────────────────────────→ │   Servidor      │
│ (Agencia)       │                                  │ (Lotería        │
│                 │                                  │  Nacional)      │
├─────────────────┤                                  ├─────────────────┤
│ - Configuración │                                  │ - Protocolo     │
│ - Protocolo     │                                  │ - Almacenamiento│
│ - Comunicación  │                                  │ - Respuestas    │
└─────────────────┘                                  └─────────────────┘
```

### Flujo de Comunicación

```
1. Cliente se conecta al servidor
2. Cliente envía: [length][agency_id|nombre|apellido|documento|nacimiento|numero]
3. Servidor recibe y parsea el mensaje
4. Servidor almacena la apuesta usando store_bets()
5. Servidor responde con el número de la apuesta como ACK
6. Cliente recibe ACK y registra éxito
7. Conexión se cierra
```

## Implementación Técnica

### 1. Protocolo de Comunicación

#### Estructura del Mensaje
```
[1 byte length][agency_id|nombre|apellido|documento|nacimiento|numero]
```

**Ejemplo de mensaje:**
```
[47][1|client1|Santiago Lionel|Lorca|30904465|1999-03-17|7574]
```

**Campos del mensaje:**
- **agency_id**: Identificador de la agencia (ej: "1")
- **nombre**: Nombre del apostador
- **apellido**: Apellido del apostador  
- **documento**: DNI del apostador
- **nacimiento**: Fecha de nacimiento (YYYY-MM-DD)
- **numero**: Número de la quiniela (4 dígitos)

#### Prevención de Short Read/Write

**Cliente (Go) - Envío:**
```go
// Envía con loop para manejar partial writes
writtenBytes := 0
for writtenBytes < len(buffer) {
    n, err := ch.conn.Write(buffer[writtenBytes:])
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    writtenBytes += n
}
```

**Cliente (Go) - Recepción:**
```go
// Lee byte por byte hasta encontrar newline
var response []byte
for {
    b, err := reader.ReadByte()
    if err != nil {
        return -1, fmt.Errorf("failed to receive message: %w", err)
    }
    response = append(response, b)
    if b == '\n' {
        break
    }
}
```

**Servidor (Python) - Envío:**
```python
# Envía con loop para manejar partial sends
total_sent = 0
while total_sent < len(message):
    sent = self.conn.send(message[total_sent:])
    total_sent += sent
```

**Servidor (Python) - Recepción:**
```python
# Lee hasta completar basado en el prefijo de longitud
while len(message_bytes) < length:
    chunk = self.conn.recv(length - len(message_bytes))
    if not chunk:
        raise ConnectionError("Connection closed by client")
    message_bytes += chunk
```

### 2. Estructura del Código

#### Cliente
```
client/
├── common/
│   └── client.go          # Lógica principal del cliente
├── protocol/
│   ├── message.go         # Definición de mensajes
│   └── communication.go   # Manejo de comunicación
├── domain/
│   └── bet.go            # Modelo de apuesta
└── main.go               # Punto de entrada
```

#### Servidor
```
server/
├── common/
│   └── server.py         # Lógica principal del servidor
├── protocol/
│   ├── message.py        # Parser de mensajes
│   └── communication.py  # Manejo de comunicación
└── main.py               # Punto de entrada
```

### 3. Manejo de Apuestas

#### Cliente
- Recibe información de apuesta desde variables de entorno
- Crea objeto `Bet` con validación básica
- Envía apuesta usando protocolo estructurado
- Espera confirmación del servidor
- Registra éxito/fallo en logs

#### Servidor
- Recibe mensaje con prefijo de longitud
- Parsea campos separados por `|`
- Crea objeto `Bet` usando `utils.Bet`
- Almacena apuesta con `store_bets([bet])`
- Responde con número de apuesta como ACK

### 4. Mejoras de Robustez

#### Manejo de Short Read/Write
- **Cliente**: Implementa loops de escritura y lectura byte-por-byte para garantizar transmisión completa
- **Servidor**: Maneja partial sends y recibe datos hasta completar el mensaje
- **Resultado**: Sistema robusto contra condiciones de red adversas

#### Optimización de Parsing
- **Uso de `strconv.Atoi()`**: Reemplaza `fmt.Sscanf()` para conversión string-to-int más eficiente
- **Parsing directo**: Elimina overhead de format strings y punteros
- **Mejor rendimiento**: Conversión más rápida y código más limpio

```go
// Antes (fmt.Sscanf)
_, err := fmt.Sscanf(response, "%d", &numericResponse)

// Después (strconv.Atoi)
numericResponse, err := strconv.Atoi(response)
```

### 5. Variables de Entorno del Cliente

```bash
CLI_ID=1                           # ID de la agencia
CLI_BET_NOMBRE=Santiago Lionel     # Nombre del apostador
CLI_BET_APELLIDO=Lorca            # Apellido del apostador
CLI_BET_DOCUMENTO=30904465        # DNI del apostador
CLI_BET_NACIMIENTO=1999-03-17    # Fecha de nacimiento
CLI_BET_NUMERO=7574               # Número de la quiniela
```

## Logs del Sistema

### Cliente
```
action: bet_created | result: success | dni: 30904465 | numero: 7574
action: apuesta_enviada | result: success | dni: 30904465 | numero: 7574
action: client_shutdown | result: success | client_id: 1
```

### Servidor
```
action: accept_connections | result: success | ip: 172.18.0.3
action: apuesta_almacenada | result: success | dni: 30904465 | numero: 7574
action: server_shutdown | result: success
```

## Características Técnicas

### Robustez de Red
- **Manejo de partial writes**: Garantiza envío completo de datos
- **Manejo de partial reads**: Asegura recepción completa de mensajes
- **Detección de desconexión**: Manejo graceful de cierres de conexión
- **Timeouts**: Prevención de bloqueos indefinidos

### Eficiencia
- **Parsing optimizado**: Uso de `strconv` para conversiones rápidas
- **Buffering inteligente**: `bufio.Reader` para lectura eficiente
- **Manejo de memoria**: Reutilización de buffers y conexiones
- **Logging estructurado**: Información detallada para debugging

