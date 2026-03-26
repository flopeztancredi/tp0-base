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
