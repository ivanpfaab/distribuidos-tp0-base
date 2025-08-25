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

**Cliente (Go):**
```go
// Envía todo en una sola operación Write
buffer := append([]byte{length}, []byte(msg)...)
conn.Write(buffer)  // Atómico: todo o nada
```

**Servidor (Python):**
```python
# Lee hasta completar basado en el prefijo de longitud
while len(message_bytes) < length:
    chunk = conn.recv(length - len(message_bytes))
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

### 4. Variables de Entorno del Cliente

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
```

### Servidor
```
action: apuesta_almacenada | result: success | dni: 30904465 | numero: 7574
```

