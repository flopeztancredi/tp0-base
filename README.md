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

## Ejercicio 5: Protocolo de Comunicación y Serialización
Se diseñó e implementó un protocolo binario para la comunicación entre las agencias (los clientes) y el servidor, prescindiendo de librerías de alto nivel (como JSON) y manejando la serialización de forma manual en Big-Endian.

### Especificación de los Mensajes

#### 1. Mensaje de Apuesta (`0x01`)
Enviado por la agencia hacia el servidor.
| Campo | Tipo | Tamaño | Descripción |
| :--- | :--- | :--- | :--- |
| **Tipo** | `uint8` | 1 byte | Valor constante `0x01`. |
| **ID Agencia** | `uint32` | 4 bytes | Identificador único de la agencia. |
| **Nombre** | `string` | 2 + N bytes | Longitud (`uint16`) seguido del texto UTF-8. |
| **Apellido** | `string` | 2 + N bytes | Longitud (`uint16`) seguido del texto UTF-8. |
| **DNI** | `uint32` | 4 bytes | Documento del apostador. |
| **Nacimiento** | `fixed` | 10 bytes | Fecha en formato `YYYY-MM-DD`. |
| **Número** | `uint32` | 4 bytes | Número apostado. |

#### 2. Mensaje de ACK (`0x02`)
Enviado por el servidor como respuesta.
| Campo | Tipo | Tamaño | Descripción |
| :--- | :--- | :--- | :--- |
| **Tipo** | `uint8` | 1 byte | Valor constante `0x02`. |
| **Estado** | `uint8` | 1 byte | `0x00` (Éxito) o `0x01` (Fallo). |

### Flujo de Comunicación
El protocolo opera bajo un esquema sincrónico:
1. **Conexión**: El cliente establece una conexión TCP por cada apuesta enviada.
2. **Serialización**: Los datos se empaquetan en un buffer de bytes para realizar un único envío (`write`), evitando la fragmentación excesiva.
3. **Confirmación**: El servidor debe responder obligatoriamente con un ACK. El cliente espera este mensaje antes de cerrar el socket para asegurar que la central recibió la apuesta.

### Detalle de Implementación: Strings Dinámicas
Los strings no tienen un tamaño fijo, sino que se envían con un prefijo de longitud:
* Se envían 2 bytes (`uint16`) con la longitud $N$ de la cadena.
* Inmediatamente después se envían los $N$ bytes del contenido.
Esto permite manejar nombres de cualquier longitud de forma segura y eficiente.

### Manejo de Short-Reads y Sincronismo
TCP es un protocolo orientado a flujo, lo que implica que los mensajes pueden llegar fragmentados. Para garantizar la integridad, se implementó la función `recv_exact`, que bloquea la lectura hasta que se haya recibido la cantidad exacta de bytes esperada para cada campo del protocolo.

## Ejercicio 6: Procesamiento por Lotes (Batching)
Se optimizó el protocolo y la lógica de envío para permitir el procesamiento de apuestas en lotes (*batches*), reduciendo el overhead de apertura/cierre de conexiones TCP y mejorando el throughput del sistema.

### Modificación del Protocolo
Se añadió un nuevo tipo de mensaje (`0x03`) para el envío de lotes:
| Campo | Tipo | Tamaño | Descripción |
| :--- | :--- | :--- | :--- |
| **Tipo** | `uint8` | 1 byte | Valor constante `0x03`. |
| **Cantidad** | `uint16` | 2 bytes | Número de apuestas contenidas en el lote ($M$). |
| **Cuerpo** | `bytes` | Variable | Concatenación de $M$ estructuras de apuesta (sin el byte de tipo individual). |

### Implementación del Cliente
* **Lectura de CSV**: El cliente lee las apuestas desde un archivo `.csv` inyectado mediante volúmenes.
* **Agrupamiento**: Se acumulan las apuestas en memoria hasta alcanzar un tamaño máximo definido por configuración (`BATCH_MAX_AMOUNT`), luego del cual se envía el lote completo al servidor en una sola conexión TCP.
* **Drenaje**: Al finalizar el archivo, se realiza un envío final con las apuestas remanentes para asegurar que no se pierda ninguna transacción.

### Implementación del Servidor
* **Atomicidad**: El servidor recibe el lote completo y persiste cada apuesta en la base de datos (archivo) de forma secuencial.
* **Respuesta**: Se envía un único ACK por cada lote procesado, simplificando el flujo de confirmación.

## Ejercicio 7: Sincronización y Sorteo
Se implementó un mecanismo de control para asegurar que el sorteo de ganadores solo se realice una vez que todas las agencias hayan finalizado la carga de sus apuestas.

### Nuevos Mensajes del Protocolo
Se agregaron tipos de mensajes para coordinar el fin de la carga y la consulta de resultados:
* **Done** (`0x04`): Enviado por el cliente para notificar que terminó de enviar todas sus apuestas.
* **Winners Request** (`0x05`): Enviado por el cliente para solicitar la lista de documentos ganadores.
* **Winners Response** (`0x06`): Enviado por el servidor con la lista de ganadores (un `uint16` para la cantidad, seguido de los DNIs como `uint32`).

### Lógica de Sincronización (Fase 1: Recolección)
El servidor opera en un ciclo de dos fases para garantizar la integridad de los datos:
1. **Espera de Agencias**: El servidor mantiene un contador de agencias activas. A medida que recibe mensajes `Done`, marca a la agencia como finalizada pero **mantiene el socket abierto**.
2. **Sorteo**: Una vez que las $N$ agencias (configuradas por `SERVER_NUM_AGENCIES`) enviaron su notificación, el servidor procede a realizar el sorteo recorriendo la base de datos de apuestas.

### Lógica de Respuesta (Fase 2: Distribución)
3. **Notificación de Ganadores**: Con el sorteo realizado, el servidor retoma los sockets en espera y responde a cada agencia con su lista particular de ganadores (aquellos documentos que pertenecen a esa agencia y salieron sorteados).

## Ejercicio 8: Multithreading y Concurrencia
Se transformó el servidor en un sistema multihilo para permitir el procesamiento simultáneo de múltiples agencias, optimizando el uso de recursos y evitando bloqueos durante la fase de carga.

### Implementación del Servidor Paralelo
* **Modelo de Hilos**: Por cada conexión entrante, el servidor crea un hilo (`threading.Thread`) que maneja la comunicación con ese cliente de forma independiente.
* **Seguridad de Datos**: Se utilizó un `threading.Lock` para sincronizar el acceso al archivo de persistencia de apuestas.
* **Sincronización de Fase**: Se implementó un `threading.Barrier` configurado con la cantidad de agencias esperadas. Los hilos que envían el mensaje `Done` quedan bloqueados en la barrera hasta que la última agencia termina.
* **Sorteo Atómico**: La barrera utiliza el parámetro `action` para ejecutar la lógica del sorteo de forma atómica y única en el momento en que se libera, asegurando que todos los hilos tengan acceso a los resultados antes de responder.
