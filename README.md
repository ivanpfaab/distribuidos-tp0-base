# TP0: Docker + Comunicaciones + Concurrencia

## Ejercicio 1: Script de Generación de Docker Compose
### Estructura del Docker Compose Generado

```yaml
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

  client1:
    container_name: client1
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=1
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

  client2:
    container_name: client2
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=2
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

  # ... más clientes según la cantidad especificada

networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
```

### Implementación Técnica

#### Generación del Archivo

El script utiliza la técnica de concantenación de texto (`cat << EOF`) para generar el contenido del archivo:

1. **Header**: Genera la sección del servidor
2. **Loop de Clientes**: Itera desde 1 hasta N para crear cada cliente
3. **Networks**: Agrega la configuración de red al final

#### Variables de Entorno
Cada cliente recibe:
- `CLI_ID`: Identificador único del cliente (1, 2, 3, ...)

### Instrucciones de Ejecución

#### 1. Verificar Permisos del Script

El script debe tener permisos de ejecución:
```bash
chmod +x generar-compose.sh
```

#### 2. Ejecutar el Script

#### Sintaxis Básica
```bash
./generar-compose.sh <archivo_salida> <cantidad_clientes>
```

#### Ejemplos de Uso

**Generar un compose con 5 clientes:**
```bash
./generar-compose.sh docker-compose-5-clients.yaml 5
```

#### 3. Verificar la Salida

El script mostrará mensajes de confirmación:
```
Generating Docker Compose file: docker-compose-5clients.yaml
Number of clients: 5
Docker Compose file generated successfully: docker-compose-5clients.yaml
The file contains 1 server and 5 client(s)
```

### Validaciones y Manejo de Errores

#### Errores Comunes

1. **Faltan Argumentos**:
   ```
   Usage: ./generar-compose.sh <output_file> <number_of_clients>
   Example: ./generar-compose.sh docker-compose-dev.yaml 5
   ```

2. **Número de Clientes Inválido**:
   ```
   Error: Number of clients must be a positive integer
   ```

3. **Archivo de Salida Existente**: El script sobrescribirá archivos existentes sin preguntar




## Ejercicio N°2:
### Diseño de la Solución

#### Arquitectura de Volúmenes

La solución implementa Docker volumes para crear un puente entre los archivos de configuración del host y los containers:

```
Host Filesystem              Container Filesystem
├── ./server/config.ini  ←→  /config.ini
└── ./client/config.yaml ←→  /config.yaml
```

### Implementación Técnica

#### 1. Modificación de Dockerfiles

#### Server Dockerfile
```dockerfile
FROM python:3.9.7-slim
COPY server /
RUN python -m unittest tests/test_common.py
ENTRYPOINT ["python3", "/main.py"]
```

**Cambios realizados:**
- **Antes**: Los archivos de configuración se copiaban durante el build
- **Después**: Los archivos se montan como volúmenes en tiempo de ejecución

#### Client Dockerfile
```dockerfile
FROM golang:1.17 AS builder
# ... build stage ...
FROM busybox:latest
COPY --from=builder /build/bin/client /client
ENTRYPOINT ["/client"]
```

**Cambios realizados:**
- **Antes**: `COPY ./client/config.yaml /config.yaml` (config en imagen)
- **Después**: Config se monta como volumen desde el host

#### 2. Configuración de Volúmenes en Docker Compose

#### Server Service
```yaml
server:
  container_name: server
  image: server:latest
  volumes:
    - ./server/config.ini:/config.ini  # Monta config.ini del host
  environment:
    - PYTHONUNBUFFERED=1
    - LOGGING_LEVEL=DEBUG
  networks:
    - testing_net
```

#### Client Service
```yaml
client1:
  container_name: client1
  image: client:latest
  volumes:
    - ./client/config.yaml:/config.yaml  # Monta config.yaml del host
  environment:
    - CLI_ID=1
    - CLI_LOG_LEVEL=DEBUG
  networks:
    - testing_net
  depends_on:
    - server
```

#### 3. Actualización del Script Generator

El script `generar-compose.sh` se actualizó para incluir volúmenes en todos los archivos generados:

### Ejemplo de Modificación en Tiempo Real

1. **Modificar configuración del servidor**:
   ```ini
   # server/config.ini
   [DEFAULT]
   SERVER_PORT = 12346  # Cambiado de 12345
   LOGGING_LEVEL = DEBUG
   ```

2. **Reiniciar solo el servidor**:
   ```bash
   make docker-compose-restart SERVICE=server
   ```

3. **Verificar cambios**:
   ```bash
   make docker-compose-logs
   ```


## Ejercicio N°3:
### Diseño de la Solución

### Arquitectura de la Solución

La solución implementa un  **contenedor temporal** que:

1. **Usa Alpine Linux como OS**
2. **Instala netcat** dentro del contenedor temporal
3. **Se conecta a la red Docker** existente (`tp0_testing_net`)
4. **Ejecuta la prueba** enviando un mensaje al servidor
5. **Captura la respuesta** y la compara con el mensaje original
6. **Limpia automáticamente** el contenedor temporal

### Implementación Técnica

#### Explicación de los componentes dentro del comando que revisa el funcionamiento del servidor

#### **Variable Assignment: `RESPONSE=`**
- **Propósito**: Captura la salida del comando completo y la almacena en la variable `RESPONSE`

#### **Command Substitution: `$()`**
- **Propósito**: Ejecuta el comando dentro de los paréntesis y retorna su salida

#### **Container Cleanup: `--rm`**
- **Propósito**: Elimina automáticamente el contenedor cuando termina la ejecución

#### **Network Connection: `--network tp0_testing_net`**
- **Propósito**: Conecta el nuevo contenedor a la red Docker existente

#### **Base Image: `alpine`**
- **Propósito**: Usa Alpine Linux como imagen base (muy pequeña, ~5MB)

#### **Shell Command: `sh -c`**
- **Propósito**: Ejecuta el shell (`sh`) con la flag `-c` para ejecutar una cadena de comando. Esto es principalmente por un tema de velocidad del test


## Ejercicio N°4:

### Diseño de la Solución

#### Arquitectura de Shutdown Graceful

1. **Captura señales** (SIGTERM/SIGINT) en ambos componentes. SIGTERM es generada por docker con el flag -t, SIGINT es para capturar el Ctrl + C en desarrollo.
2. **Se establecen flags de "running"** para coordinar la terminación y entender cuando un proceso esta corriendo y cuando se debe terminar.


#### **Flujo de Señales**

1. **Docker Compose Down**: Envía SIGTERM a todos los contenedores
2. **Signal Handlers**: Capturan SIGTERM y establecen flags de shutdown
3. **Main Loops**: Detectan flags y cortan el loop
4. **Resource Cleanup**: Se ejecuta cleanup de recursos

#### **Timeout y Force Kill**

```bash
# Graceful shutdown con timeout de 10 segundos
docker compose down -t 10

# Después del timeout, Docker envía SIGKILL si es necesario
```

### Probar Shutdown Graceful

#### **Shutdown Normal**
```bash
# Iniciar servicios
make docker-compose-up

# Shutdown graceful
make docker-compose-down
```

#### **Shutdown con Timeout**
```bash
# Shutdown con timeout específico
docker compose down -t 10
```

### Logs de Shutdown

#### **Servidor Graceful Shutdown**
```
server   | action: signal_received | result: success | signal: 15
server   | action: server_shutdown | result: in_progress
server   | action: cleanup_resources | result: in_progress
server   | action: cleanup_resources | result: success | resource: server_socket
server   | action: cleanup_resources | result: success
server   | action: server_shutdown | result: success
```

#### **Cliente Graceful Shutdown**
```
client1   | action: signal_received | result: success | client_id: 1 | signal: terminated
client1   | action: client_shutdown | result: in_progress | client_id: 1
client1   | action: cleanup_resources | result: in_progress | client_id: 1
client1   | action: cleanup_resources | result: success | client_id: 1 | resource: connection
client1   | action: cleanup_resources | result: success | client_id: 1
client1   | action: client_shutdown | result: success | client_id: 1
```

## Ejercicio 5

### Diseño de la Solución

#### Arquitectura del Sistema

```
┌─────────────────┐    Protocolo de Comunicación     ┌─────────────────┐
│   Cliente       │ ←──────────────────────────────→ │   Servidor      │
│ (Agencia)       │                                  │ (Lotería        │
│                 │                                  │  Nacional)      │
├─────────────────┤                                  ├─────────────────┤
│                 │                                  │ - Protocolo     │
│ - Protocolo     │                                  │ - Comunicación  │
│ - Comunicación  │                                  │                 │
└─────────────────┘                                  └─────────────────┘
```

#### Flujo de Comunicación

```
1. Cliente se conecta al servidor
2. Cliente envía: [length][agency_id|nombre|apellido|documento|nacimiento|numero]
3. Servidor recibe y parsea el mensaje
4. Servidor almacena la apuesta usando store_bets()
5. Servidor responde con el número de la apuesta como ACK
6. Cliente recibe ACK y registra éxito
7. Conexión se cierra
```

### Implementación Técnica

#### Protocolo de Comunicación

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

#### Variables de Entorno del Cliente

```bash
CLI_ID=1                           # ID de la agencia
CLI_BET_NOMBRE=Santiago Lionel     # Nombre del apostador
CLI_BET_APELLIDO=Lorca            # Apellido del apostador
CLI_BET_DOCUMENTO=30904465        # DNI del apostador
CLI_BET_NACIMIENTO=1999-03-17    # Fecha de nacimiento
CLI_BET_NUMERO=7574               # Número de la quiniela
```

### Logs del Sistema

#### Cliente
```
action: bet_created | result: success | dni: 30904465 | numero: 7574
action: apuesta_enviada | result: success | dni: 30904465 | numero: 7574
action: client_shutdown | result: success | client_id: 1
```

#### Servidor
```
action: accept_connections | result: success | ip: 172.18.0.3
action: apuesta_almacenada | result: success | dni: 30904465 | numero: 7574
action: server_shutdown | result: success
```

## Ejercicio 6

### Diseño de la Solución

#### Arquitectura del Sistema de Lotes

```
┌─────────────────┐    Batch   Protocol        ┌─────────────────┐
│   Cliente       │ ←────────────────────────→ │   Servidor      │
│ (Agencia)       │                            │ (Lotería        │
│                 │                            │  Nacional)      │
├─────────────────┤                            ├─────────────────┤
│ - CSV Reader    │                            │ - Parser de     │
│ - Chunk Reader  │                            │  Lotes          │
│ - Batch Sender  │                            │ - Procesador    │
└─────────────────┘                            └─────────────────┘
```

#### Flujo de Procesamiento por batchs

```
1. Cliente lee archivo CSV de apuestas
2. Cliente agrupa apuestas en chunks configurables
3. Cliente envía lote completo al servidor
4. Servidor recibe y parsea el lote completo
5. Servidor procesa todas las apuestas del lote
6. Servidor responde con confirmación del lote
7. Cliente continúa con el siguiente lote
```

### Limitaciones de Tamaño de Paquete

#### Estructura del Mensaje de Batch

El sistema utiliza un protocolo de batches con la siguiente estructura:

```
<msg_size_8_bytes><bet1>&&<bet2>&&<bet3>&&...
```

Donde:
- `msg_size_8_bytes`: 8 bytes para el tamaño total del contenido
- Cada apuesta: `<agency_id>|<nombre>|<apellido>|<documento>|<nacimiento>|<numero>&&`
- `&&`: Separador entre apuestas

### Cálculo del Tamaño por Apuesta

**Campos de una apuesta típica:**
- `agency_id`: 1-2 dígitos = 1-2 bytes
- `|`: 1 byte (separador)
- `nombre`: ~10-15 caracteres = 10-15 bytes
- `|`: 1 byte
- `apellido`: ~10-15 caracteres = 10-15 bytes
- `|`: 1 byte
- `documento`: 8 dígitos = 8 bytes
- `|`: 1 byte
- `nacimiento`: YYYY-MM-DD = 10 bytes
- `|`: 1 byte
- `numero`: 4 dígitos = 4 bytes
- `&&`: 2 bytes (separador de apuesta)

**Total por apuesta**: ~40-50 bytes

#### Máximo de Apuestas en 8KB

```
8KB = 8192 bytes
- 8 bytes (prefijo de tamaño) = 8184 bytes disponibles
- 8184 bytes ÷ 45 bytes por apuesta (promedio) = ~181 apuestas
- 8184 bytes ÷ 40 bytes por apuesta (mínimo) = ~204 apuestas
- 8184 bytes ÷ 50 bytes por apuesta (máximo) = ~163 apuestas
```

#### Recomendación Práctica

**Máximo recomendado: 160 apuestas por lote**

Esta limitación asegura que el paquete permanezca bajo 8KB en todos los casos, considerando:
- Nombres largos y variables
- Margen de seguridad para variaciones
- Eficiencia en el procesamiento de batchs

### Configuración del Tamaño de batch

```yaml
# client/config.yaml
batch:
  maxAmount: 160  
```

## Logs del Sistema

### Cliente
```
action: open_file | result: success | client_id: 1
action: read_chunk | result: success | client_id: 1 | chunk_size: 99
action: batch_sent | result: success | cantidad: 99
action: read_chunk | result: success | client_id: 1 | chunk_size: 99
action: batch_sent | result: success | cantidad: 99
action: loop_finished | result: success | client_id: 1
```

### Servidor
```
action: apuesta_recibida | result: success | cantidad: 99
action: apuesta_recibida | result: success | cantidad: 99
action: client_disconnected | result: success | detail: client finished sending data
```

## Instrucciones de Ejecución

### 1. Preparar Archivos de Datos
```bash
# Verificar que los archivos CSV estén en .data/
ls -la .data/agency-*.csv
```

### 2. Configurar Tamaño de Lote
```yaml
# Modificar client/config.yaml
batch:
  maxAmount: 99  # Ajustar según necesidades
```

### 3. Ejecutar Sistema
```bash
# Iniciar servicios
make docker-compose-up

# Ver logs
make docker-compose-logs

# Detener servicios
make docker-compose-down
```

### 4. Verificar Funcionamiento
```bash
# Ver logs del cliente
make docker-compose-logs | grep client1

# Ver logs del servidor
make docker-compose-logs | grep server
```
## Ejercicio 7

### Arquitectura de la Solución

La solución implementa un **protocolo de comunicación** que:

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

### Implementación Técnica

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

### 2. Flujo de Interacción

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

### Características de la Solución

#### 1. **Protocolo con Tipos de Mensaje**
- **Prefijo de tipo**: Identificación clara del tipo de mensaje
- **Prefijo de tamaño**: Manejo seguro de mensajes de cualquier longitud
- **Separación de responsabilidades**: Cada tipo de mensaje tiene un propósito específico

#### 2. **Tracking Dinámico de Clientes**
- **No asume número fijo**: El servidor aprende dinámicamente cuántas agencias existen
- **Conexiones activas**: Rastrea conexiones actualmente abiertas
- **Agencias completadas**: Mantiene registro de qué agencias han notificado finalización
- **Sincronización**: Solo procede con el sorteo cuando todas las agencias activas han completado

#### 3. **Estrategia de Desconexión/Reconexión**
- **Procesamiento secuencial**: Permite que múltiples clientes procesen sin conflictos
- **Liberación de recursos**: Desconecta después de enviar apuestas para liberar conexiones
- **Reconexión inteligente**: Se reconecta solo para consultar ganadores
- **Retry con backoff**: Implementa reintentos con delay exponencial para consultas de ganadores

#### 4. **Manejo de Estados del Servidor**
- **Estado de lotería**: Controla si el sorteo ya se realizó
- **Cache de ganadores**: Almacena ganadores agrupados por agencia
- **Respuestas condicionales**: Responde "waiting" si no todas las agencias han completado
- **Integridad de datos**: Solo permite consultas después de que todas las agencias hayan terminado


### Instrucciones de Ejecución

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

### Ejemplo de Logs Esperados

#### Cliente
```
client1  | action: batch_sent | result: success | cantidad: 4
client1  | action: notification_sent | result: success | client_id: 1
client1  | action: waiting_for_other_clients | result: success | client_id: 1
client1  | action: winner_query | result: waiting | client_id: 1 | attempt: 1
client1  | action: consulta_ganadores | result: success | cant_ganadores: 2
client1  | action: client_shutdown | result: success | client_id: 1
```

#### Servidor
```
server   | action: new_client_connected | result: success | client_number: 172.18.0.3
server   | action: apuesta_recibida | result: success | cantidad: 4
server   | action: agency_completed | result: success | agency_id: 1
server   | action: client_sent_all_bets | result: success | client_id: 172.18.0.3
server   | action: sorteo | result: success
server   | action: winner_query | result: success | agency_id: 1 | winners_count: 2
server   | action: winners_sent | result: success | winners_count: 2
```

## Ejercicio 8

### Diseño de la Solución

#### Arquitectura de Concurrencia

```
┌─────────────────┐    Múltiples Conexiones      ┌─────────────────┐
│   Cliente 1     │ ←──────────────────────────→ │                 │
│   Cliente 2     │ ←──────────────────────────→ │   Servidor      │
│   Cliente 3     │ ←──────────────────────────→ │   Multithreaded │
│   Cliente N     │ ←──────────────────────────→ │                 │
└─────────────────┘                              └─────────────────┘
                                                          │
                                                          ▼
                                                 ┌─────────────────┐
                                                 │                 │
                                                 │ - Thread 1      │
                                                 │ - Thread 2      │
                                                 │ - Thread N      │
                                                 └─────────────────┘
```

### Modelo de Threading

```
Main Thread (Accept Loop)
    ├── Thread 1 (Client 1)
    ├── Thread 2 (Client 2)
    ├── Thread 3 (Client 3)
    └── Thread N (Client N)
```

## Implementación Técnica

### 1. Sistema de Locks Thread-Safe

#### Definición de Locks
```python
class Server:
    def __init__(self, port, listen_backlog):
        # Thread-safe locks for shared resources
        self._completion_lock = threading.Lock()    # Protege _completed_agencies
        self._lottery_lock = threading.Lock()       # Protege operaciones de lotería
        self._storage_lock = threading.Lock()       # Protege almacenamiento de apuestas
        self._connections_lock = threading.Lock()   # Protege _active_connections
        self._threads_lock = threading.Lock()       # Protege _active_threads
```

#### Uso de Locks en Operaciones Críticas

**Almacenamiento de Apuestas (Thread-Safe):**
```python
def __handle_batch_bets(self, communication_handler, bets):
    try:
        bets_stored = 0
        
        # Proteger el almacenamiento con lock
        with self._storage_lock:
            for bet in bets:
                if bet:
                    store_bets([bet])  # Operación crítica
                    bets_stored += 1
        
        communication_handler.send_batch_response(len(bets), bets_stored)
    except Exception as e:
        logging.error(f'action: handle_batch_bets | result: fail | error: {e}')
```

**Manejo de Notificaciones (Thread-Safe):**
```python
def __handle_notification(self, communication_handler, agency_id):
    try:
        # Proteger acceso a _completed_agencies
        with self._completion_lock:
            self._completed_agencies.add(agency_id)
            logging.info(f'action: agency_completed | result: success | agency_id: {agency_id}')
        
        communication_handler.send_notification_response(True)
    except Exception as e:
        logging.error(f'action: handle_notification | result: fail | agency_id: {agency_id} | error: {e}')
```

**Consulta de Ganadores (Thread-Safe):**
```python
def __handle_winner_query(self, communication_handler, agency_id):
    try:
        # Verificar estado de manera thread-safe
        with self._completion_lock:
            completed_count = len(self._completed_agencies)
        with self._connections_lock:
            active_count = len(self._active_connections)
        
        if completed_count < active_count:
            communication_handler.send_waiting_response()
            return
        
        # Proteger operación de lotería
        with self._lottery_lock:
            if not self._lottery.is_lottery_conducted():
                self._lottery.conduct_lottery()
        
        winners = self._lottery.get_winners_for_agency(agency_id)
        communication_handler.send_winner_list(winners)
    except Exception as e:
        logging.error(f'action: handle_winner_query | result: fail | agency_id: {agency_id} | error: {e}')
```

### 2. Creación y Gestión de Threads

#### Creación de Threads por Cliente
```python
def run(self):
    """Main server loop that accepts connections and creates a thread for each"""
    logging.info('action: server_start | result: success')
    
    try:
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                if client_sock:
                    # Crear thread dedicado para cada cliente
                    client_thread = threading.Thread(
                        target=self.__handle_client_connection, 
                        args=(client_sock,),
                        daemon=False,  # Non-daemon para poder esperar
                    )
                    
                    # Registrar thread para shutdown graceful
                    with self._threads_lock:
                        self._active_threads.add(client_thread)
                    
                    client_thread.start()
                    
            except socket.timeout:
                continue
            except Exception as e:
                logging.error(f"action: accept_connection | result: fail | error: {e}")
                continue
    finally:
        self.__cleanup_resources()
```

#### Gestión de Conexiones Thread-Safe
```python
def __accept_new_connection(self):
    """Accept new connections and track them thread-safely"""
    logging.info('action: accept_connections | result: in_progress')
    c, addr = self._server_socket.accept()
    logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
    
    # Registrar conexión de manera thread-safe
    with self._connections_lock:
        self._active_connections.add(c)
    
    return c
```


### Flujo de Ejecución Concurrente

#### Timeline de Threads
```
T0: Main Thread inicia, acepta conexiones
T1: Cliente 1 conecta → Thread 1 creado
T2: Cliente 2 conecta → Thread 2 creado
T3: Cliente 3 conecta → Thread 3 creado
T4: Thread 1 procesa apuestas de Cliente 1
T5: Thread 2 procesa apuestas de Cliente 2 (paralelo)
T6: Thread 3 procesa apuestas de Cliente 3 (paralelo)
T7: Thread 1 notifica finalización
T8: Thread 2 notifica finalización
T9: Thread 3 notifica finalización
T10: Thread 1 consulta ganadores
T11: Thread 2 consulta ganadores (paralelo)
T12: Thread 3 consulta ganadores (paralelo)
T13: Shutdown → Todos los threads terminan
```

#### Estados de Sincronización
```
Estado Inicial: Main Thread esperando conexiones
Estado Procesando: Múltiples threads procesando clientes
Estado Sincronización: Locks protegiendo recursos compartidos
Estado Lotería: Un solo thread puede ejecutar lotería
Estado Consultas: Múltiples threads consultando ganadores
Estado Shutdown: Todos los threads terminando ordenadamente
```

## Logs del Sistema Concurrente

### Servidor
```
server   | action: server_start | result: success
server   | action: accept_connections | result: success | ip: 172.18.0.3
server   | action: new_client_connected | result: success | client_number: 172.18.0.3
server   | action: accept_connections | result: success | ip: 172.18.0.4
server   | action: new_client_connected | result: success | client_number: 172.18.0.4
server   | action: apuesta_recibida | result: success | cantidad: 4
server   | action: apuesta_recibida | result: success | cantidad: 3
server   | action: agency_completed | result: success | agency_id: 1
server   | action: agency_completed | result: success | agency_id: 2
server   | action: sorteo | result: success
server   | action: winner_query | result: success | agency_id: 1 | winners_count: 2
server   | action: winner_query | result: success | agency_id: 2 | winners_count: 1
server   | action: cleanup_resources | result: waiting_for_threads | active: 2
server   | action: cleanup_resources | result: thread_finished | thread: Thread-1
server   | action: cleanup_resources | result: thread_finished | thread: Thread-2
server   | action: server_shutdown | result: success
```