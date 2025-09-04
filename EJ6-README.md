# Ejercicio 6: Procesamiento por Lotes (Batch Processing) de Apuestas

## Diseño de la Solución

### Arquitectura del Sistema de Lotes

```
┌─────────────────┐    Protocolo de Lotes      ┌─────────────────┐
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

## Limitaciones de Tamaño de Paquete

### Estructura del Mensaje de Lote

El sistema utiliza un protocolo de lotes con la siguiente estructura:

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

### Máximo de Apuestas en 8KB

```
8KB = 8192 bytes
- 8 bytes (prefijo de tamaño) = 8184 bytes disponibles
- 8184 bytes ÷ 45 bytes por apuesta (promedio) = ~181 apuestas
- 8184 bytes ÷ 40 bytes por apuesta (mínimo) = ~204 apuestas
- 8184 bytes ÷ 50 bytes por apuesta (máximo) = ~163 apuestas
```

### Recomendación Práctica

**Máximo recomendado: 160-170 apuestas por lote**

Esta limitación asegura que el paquete permanezca bajo 8KB en todos los casos, considerando:
- Nombres largos y variables
- Margen de seguridad para variaciones
- Eficiencia en el procesamiento de lotes

### Configuración del Tamaño de Lote

```yaml
# client/config.yaml
batch:
  maxAmount: 160  # Ajustar según necesidades (máximo recomendado: 170)
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