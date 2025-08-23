# Ejercicio 3: Script de Validación del Echo Server

## Problema a Resolver

### Objetivo
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo.

### Requisitos Específicos
- **Funcionalidad**: Enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado
- **Output**: Imprimir `action: test_echo_server | result: success` en caso exitoso, o `action: test_echo_server | result: fail` en caso contrario
- **Restricciones**: 
  - Netcat no debe ser instalado en la máquina host
  - No se pueden exponer puertos del servidor para realizar la comunicación
  - Debe utilizar la red Docker para la comunicación

## Diseño de la Solución

### Arquitectura de la Solución

La solución implementa un  **contenedor temporal** que:

1. **Usa Alpine Linux como OS**
2. **Instala netcat** dentro del contenedor temporal
3. **Se conecta a la red Docker** existente (`tp0_testing_net`)
4. **Ejecuta la prueba** enviando un mensaje al servidor
5. **Captura la respuesta** y la compara con el mensaje original
6. **Limpia automáticamente** el contenedor temporal

## Implementación Técnica

### Explicación de los componentes dentro del comando

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

## Pruebas y descubrimientos observados a lo largo del desarrollo

### **¿Por qué no puedo usar un cliente existente?**

#### **El Problema con Contenedores Cliente Existentes**

Los contenedores cliente están diseñados con un **ciclo de vida finito**:

- Se conectan al servidor
- Envían un número predeterminado de mensajes (5 en tu caso)
- Salen automáticamente cuando terminan

#### **Por qué `docker exec` Falla en contenedores levantados con el YAML**

El comando `docker exec` solo puede ejecutarse en **contenedores en ejecución**. Cuando un contenedor sale:

```bash
# Esto fallará si client1 ya esta cerrado
docker exec client1 sh -c "echo 'test' | nc server 12345"
```

### **¿Por qué la red se llama `tp0_testing_net` y no `testing_net`?**

#### **Convención de Nomenclatura de Docker Compose**

Cuando se define una red en el archivo `docker-compose-dev.yaml`:

```yaml
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
```

Docker Compose automáticamente prefija el nombre de la red con el **nombre del proyecto** y un guión bajo.

## Instrucciones de Ejecución

### 1. Verificar Permisos del Script

El script debe tener permisos de ejecución:
```bash
chmod +x validar-echo-server.sh
```

### 2. Ejecutar el Script
```bash
./validar-echo-server.sh
```

#### Prerrequisitos
- Docker debe estar ejecutándose
- Los contenedores del proyecto deben estar activos
- Ejecutar `make docker-compose-up` antes de usar el script

### 3. Verificar la Salida

#### Caso Exitoso
```
action: test_echo_server | result: success
```

#### Caso de Falla
```
action: test_echo_server | result: fail
```