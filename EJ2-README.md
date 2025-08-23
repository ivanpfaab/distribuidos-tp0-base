# Ejercicio 2: Configuración Persistente con Docker Volumes

## Problema a Resolver

### Antes de la Solución
- Los archivos de configuración (`config.ini` y `config.yaml`) se copiaban dentro de las imágenes Docker durante el build
- Cualquier cambio en la configuración requería:
  1. Modificar el archivo de configuración
  2. Reconstruir la imagen Docker completa
  3. Reiniciar el container
- **Resultado**: Proceso lento y poco eficiente para desarrollo

### Después de la Solución
- Los archivos de configuración se montan como volúmenes desde el host
- **Resultado**: Proceso rápido y eficiente, sin necesidad de rebuilds

## Diseño de la Solución

### Arquitectura de Volúmenes

La solución implementa Docker volumes para crear un puente entre los archivos de configuración del host y los containers:

```
Host Filesystem              Container Filesystem
├── ./server/config.ini  ←→  /config.ini
└── ./client/config.yaml ←→  /config.yaml
```

## Implementación Técnica

### 1. Modificación de Dockerfiles

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

### 2. Configuración de Volúmenes en Docker Compose

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

### 3. Actualización del Script Generator

El script `generar-compose.sh` se actualizó para incluir volúmenes en todos los archivos generados:

## Ejemplo de Modificación en Tiempo Real

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