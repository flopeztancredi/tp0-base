# Resolución TP0 - Sistemas Distribuidos I

## Ejercicio 1: Generación Dinámica de Docker Compose
Se implementó el script `generar-compose.sh` para automatizar la creación del entorno de servicios. 

### Uso
```bash
./generar-compose.sh <archivo_salida> <cantidad_clientes>
```

### Ejemplo
Para generar un entorno con 5 clientes:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

### Implementación
* **Servicio Server**: Define un contenedor único de Python.
* **Servicios Client**: Genera $N$ contenedores de Go (`client1` a `clientN`) mediante un bucle bash.
* **Configuración**: Inyecta la variable de entorno `CLI_ID` de forma incremental para identificar cada cliente.
* **Redes**: Configura la red `testing_net` (bridge) con la subred `172.25.125.0/24` para permitir la comunicación por nombre de host (`server`).

## Ejercicio 2: Configuración a través de Volúmenes
Se implementaron volúmenes para montar archivos de configuración desde el host a los contenedores, permitiendo realizar cambios en la configuración sin necesidad de reconstruir las imágenes.

### Uso

```bash
make docker-compose-up
```

Al ejecutar este comando, los servicios se inician utilizando los archivos de configuración:
* `./server/config.ini` para el servicio `server`.
* `./client/config.yaml` para los servicios `client`.

### Implementación

* **Montaje de Volúmenes**: Se configuraron volúmenes en el `docker-compose-dev.yaml` para mapear los archivos de configuración desde el host a los contenedores.

## Ejercicio 3: Networking y Validación de Conectividad
Se implementó el script `validar-echo-server.sh` para verificar la disponibilidad del servidor sin exponer puertos al host externo, utilizando la red interna de Docker.

### Uso
```bash
./validar-echo-server.sh
```
El script envía un mensaje al servidor y valida que la respuesta sea idéntica al envío (Echo Server).

### Notas de Implementación
* **Contenedor Temporal**: El script lanza un contenedor de `alpine:latest` unido a la red `tp0_testing_net`.
* **Comunicación Interna**: Se utiliza el comando `nc` (Netcat) para enviar el mensaje directamente al host `server` en el puerto `12345`.
* **Resultado**: El script reporta `success` si los mensajes coinciden, o `fail` en caso contrario o por timeout.

## Ejercicio 4: Graceful Shutdown
Se implementó el manejo de señales del sistema operativo para asegurar un cierre limpio de los recursos (sockets, archivos y descriptores) ante una interrupción.

### Uso
Al ejecutar `make docker-compose-down` o enviar una señal `SIGTERM` (Ctrl+C en la terminal de logs), los servicios registrarán los logs de cierre y finalizarán correctamente.

### Ejemplo de Graceful Shutdown
Al detener el entorno con `make docker-compose-down`, se puede observar en los logs la captura de la señal y el cierre de recursos:

```bash
server   | 2026-03-26 03:25:49 INFO     action: graceful_shutdown | result: in_progress
server   | 2026-03-26 03:25:49 INFO     action: graceful_shutdown | result: success
client1  | 2026-03-26 03:25:47 INFO     action: graceful_shutdown | result: in_progress | client_id: 1
client1  | 2026-03-26 03:25:47 INFO     action: graceful_shutdown | result: success | client_id: 1
```

### Notas de Implementación
* **Servidor (Python)**: Se utilizó el módulo `signal` para capturar `SIGTERM`. Al recibirse la señal, el servidor cierra el socket principal de escucha, lo que rompe el bloqueo del `accept()` y permite una salida controlada del bucle principal.
* **Cliente (Go)**: Se implementó mediante `signal.NotifyContext` de la librería estándar. El ciclo de vida del cliente está atado a un contexto que se cancela al recibir la señal, permitiendo cerrar la conexión activa y abortar iteraciones pendientes mediante un `select`.
* **Logs**: Se añadieron mensajes con el formato `action: graceful_shutdown | result: in_progress/success` para validar la secuencia de cierre en ambos lenguajes.
