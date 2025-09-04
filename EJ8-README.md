# Ejercicio 8: Procesamiento Concurrente con Multithreading

## Problema a Resolver

### Objetivo
Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo, implementando un sistema de concurrencia que maneje múltiples clientes simultáneamente.

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