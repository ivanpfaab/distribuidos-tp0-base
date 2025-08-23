# Ejercicio 1: Script de Generación de Docker Compose

## Estructura del Docker Compose Generado

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

## Implementación Técnica

### Generación del Archivo

El script utiliza la técnica de concantenación de texto (`cat << EOF`) para generar el contenido del archivo:

1. **Header**: Genera la sección del servidor
2. **Loop de Clientes**: Itera desde 1 hasta N para crear cada cliente
3. **Networks**: Agrega la configuración de red al final

### Variables de Entorno
Cada cliente recibe:
- `CLI_ID`: Identificador único del cliente (1, 2, 3, ...)

## Instrucciones de Ejecución

## 1. Verificar Permisos del Script

El script debe tener permisos de ejecución:
```bash
chmod +x generar-compose.sh
```

## 2. Ejecutar el Script

#### Sintaxis Básica
```bash
./generar-compose.sh <archivo_salida> <cantidad_clientes>
```

#### Ejemplos de Uso

**Generar un compose con 5 clientes:**
```bash
./generar-compose.sh docker-compose-3clients.yaml 5
```

## 3. Verificar la Salida

El script mostrará mensajes de confirmación:
```
Generating Docker Compose file: docker-compose-5clients.yaml
Number of clients: 5
Docker Compose file generated successfully: docker-compose-5clients.yaml
The file contains 1 server and 5 client(s)
```

## Validaciones y Manejo de Errores

### Errores Comunes

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