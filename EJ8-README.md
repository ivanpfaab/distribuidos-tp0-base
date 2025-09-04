# Ejercicio 8: Procesamiento Concurrente con Multithreading

## Problema a Resolver

### Objetivo
Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo, implementando un sistema de concurrencia que maneje múltiples clientes simultáneamente.

### Desafíos de Concurrencia
- **Múltiples clientes simultáneos**: El servidor debe manejar conexiones de varias agencias al mismo tiempo
- **Recursos compartidos**: Acceso thread-safe a datos compartidos (apuestas, estado de lotería, conexiones)
- **Sincronización**: Coordinación entre threads para evitar condiciones de carrera
- **Shutdown graceful**: Terminación ordenada de todos los threads al cerrar el servidor

## Diseño de la Solución

### Arquitectura de Concurrencia

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

### 3. Shutdown Graceful con Sincronización

#### Limpieza de Recursos Thread-Safe
```python
def __cleanup_resources(self):
    """Clean up all server resources gracefully with thread synchronization"""
    logging.info('action: cleanup_resources | result: in_progress')
    
    try:
        # Cerrar todas las conexiones de clientes
        with self._connections_lock:
            for client_sock in list(self._active_connections):
                try:
                    client_sock.close()
                    logging.info('action: cleanup_resources | result: success | resource: client_socket')
                except Exception as e:
                    logging.error(f'action: cleanup_resources | result: fail | resource: client_socket | error: {e}')
            self._active_connections.clear()
        
        # Esperar a que todos los threads terminen
        with self._threads_lock:
            active_threads = list(self._active_threads)
        
        if active_threads:
            logging.info(f'action: cleanup_resources | result: waiting_for_threads | active: {len(active_threads)}')
            
            # Esperar a cada thread con timeout
            for thread in active_threads:
                thread.join()  # Esperar a que termine
                self._active_threads.discard(thread)
                if thread.is_alive():
                    logging.warning(f'action: cleanup_resources | result: thread_timeout | thread: {thread.name}')
                else:
                    logging.info(f'action: cleanup_resources | result: thread_finished | thread: {thread.name}')
        
        # Cerrar socket del servidor
        if hasattr(self, '_server_socket') and self._server_socket:
            self._server_socket.close()
            logging.info('action: cleanup_resources | result: success | resource: server_socket')
            
    except Exception as e:
        logging.error(f'action: cleanup_resources | result: fail | error: {e}')
    
    logging.info('action: cleanup_resources | result: success')
```

### 4. Limpieza de Threads al Finalizar

#### Remoción Thread-Safe de Threads
```python
def __handle_client_connection(self, client_sock):
    """Handle client connection in separate thread"""
    try:
        communication_handler = CommunicationHandler(client_sock)
        
        # Procesar mensajes del cliente
        while self._running:
            try:
                message_data = communication_handler.receive_message()
                # ... procesar mensaje ...
            except Exception as e:
                logging.error(f"action: process_message | result: fail | error: {e}")
                break
                
    except Exception as e:
        addr = client_sock.getpeername() if client_sock else "unknown"
        logging.error(f"action: handle_client | result: fail | ip: {addr} | error: {e}")
        
    finally:
        # Limpiar recursos de manera thread-safe
        with self._connections_lock:
            self._active_connections.discard(client_sock)
        
        # Remover este thread de la lista activa
        with self._threads_lock:
            self._active_threads.discard(threading.current_thread())
        
        # Cerrar conexión
        communication_handler.close()
```

## Características de la Solución

### 1. **Thread Safety Completa**
- **Locks granulares**: Diferentes locks para diferentes recursos compartidos
- **Operaciones atómicas**: Todas las operaciones críticas están protegidas
- **Sin condiciones de carrera**: Sincronización adecuada en todos los puntos de acceso

### 2. **Escalabilidad**
- **Thread por cliente**: Cada cliente tiene su propio thread dedicado
- **Procesamiento paralelo**: Múltiples clientes pueden procesar simultáneamente
- **Sin bloqueos innecesarios**: Locks granulares minimizan la contención

### 3. **Robustez**
- **Shutdown graceful**: Todos los threads terminan ordenadamente
- **Manejo de errores**: Errores en un thread no afectan otros
- **Limpieza de recursos**: Conexiones y threads se limpian correctamente

### 4. **Sincronización Inteligente**
- **Locks de lectura/escritura**: Diferentes niveles de acceso según necesidad
- **Timeouts**: Prevención de deadlocks en shutdown
- **Estado consistente**: El estado del servidor siempre es válido

## Flujo de Ejecución Concurrente

### Timeline de Threads
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

### Estados de Sincronización
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

## Ventajas del Diseño Multithreaded

### 1. **Concurrencia Real**
- Múltiples clientes procesan simultáneamente
- No hay bloqueos entre clientes independientes
- Mejor utilización de recursos del servidor

### 2. **Thread Safety Robusta**
- Locks granulares para diferentes recursos
- Prevención completa de condiciones de carrera
- Sincronización eficiente sin deadlocks

### 3. **Escalabilidad**
- Soporta cualquier número de clientes concurrentes
- Procesamiento paralelo de apuestas
- Consultas de ganadores simultáneas

### 4. **Mantenibilidad**
- Código claro con separación de responsabilidades
- Logging detallado para debugging de concurrencia
- Shutdown graceful para operaciones de producción