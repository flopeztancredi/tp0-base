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

### Implementación

* **Contenedor Temporal**: El script lanza un contenedor de `alpine:latest` unido a la red `tp0_testing_net`.
* **Comunicación Interna**: Se utiliza el comando `nc` (Netcat) para enviar el mensaje directamente al host `server` en el puerto `12345`.
* **Resultado**: El script reporta `success` si los mensajes coinciden, o `fail` en caso contrario o por timeout.
