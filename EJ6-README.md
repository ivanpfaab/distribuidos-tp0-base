# Ejercicio 6: Procesamiento por Lotes (Batch Processing) de Apuestas

## Diseño de la Solución

### Arquitectura del Sistema de Lotes

```
┌─────────────────┐    Protocolo de Lotes     ┌─────────────────┐
│   Cliente       │ ←────────────────────────→ │   Servidor      │
│ (Agencia)       │                            │ (Lotería        │
│                 │                            │  Nacional)      │
├─────────────────┤                            ├─────────────────┤
│ - CSV Reader    │                            │ - Parser de     │
│ - Chunk Reader  │                            │  Lotes          │
│ - Batch Sender  │                            │ - Procesador    │
└─────────────────┘                            └─────────────────┘
```

### Flujo de Procesamiento por Lotes

```
1. Cliente lee archivo CSV de apuestas
2. Cliente agrupa apuestas en chunks configurables
3. Cliente envía lote completo al servidor
4. Servidor recibe y parsea el lote completo
5. Servidor procesa todas las apuestas del lote
6. Servidor responde con confirmación del lote
7. Cliente continúa con el siguiente lote
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

