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
