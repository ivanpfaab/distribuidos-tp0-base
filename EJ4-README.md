# Ejercicio 4: Terminación Graceful con SIGTERM

## Problema a Resolver

### Objetivo
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM.

## Diseño de la Solución

### Arquitectura de Shutdown Graceful

1. **Captura señales** (SIGTERM/SIGINT) en ambos componentes. SIGTERM es generada por docker con el flag -t, SIGINT es para capturar el Ctrl + C en desarrollo.
2. **Se establecen flags de "running"** para coordinar la terminación y entender cuando un proceso esta corriendo y cuando se debe terminar.

## Implementación

### Coordinación de Shutdown

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

### 3. Verificar Logs de Shutdown

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
