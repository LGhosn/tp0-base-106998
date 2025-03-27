# 7574 Sistemas Distribuidos - TP0
## Lautaro Gabriel Ghosn - 106998

## Parte 1: Introducción a Docker

### Ejercicio 1
Se creó el archivo `generar-compose.sh`, el cual define la estructura base de un archivo Docker Compose que incluye un servidor, una red y n clientes. El script recibe como parámetros el nombre del archivo a generar y la cantidad de clientes que se desean crear.

#### Ejecucion

``` bash
./generar-compose.sh <archivo de salida> <cantidad de clientes>
```

### Ejercicio 2
Al archivo creado en el anterior ejercicio se le agrega la funcionalidad meidante el uso de `docker volumes` para que tanto el server como los clientes extraigan la información predeterminada en archivos de configuración.

### Ejercicio 3
En este punto se crea un nuevo script `validar-echo-server.sh` que permite verificar si el server esta funcionando correctamente.
Se utilizó la imagen `subfuzion/netcat` para poder conectar con el servidor y verificar que su funcionamiento sea el esperado.

#### Ejecucion

``` bash
./validar-echo-server.sh
```

### Ejercicio 4
Se modificaron el servidor y el cliente para que, al recibir la señal `SIGTERM`, la capturen y cierren adecuadamente todos los recursos abiertos, evitando posibles fugas de memoria y garantizando un apagon ordenado (`graceful shutdown`).

## Parte 2: Repaso de Comunicaciones

### Ejericio 5
En este ejercicio se realizó la primera iteracion sobre la comunicación entre el cliente y el server. Se implementó la escritura y lectura de lado del cliente pero del lado del servidor solo la escritura ya que no la requería a este punto.

El cliente primero envía el largo bytes de la apuesta que está por mandar.
El size está codificado como un uint32 por lo que se envían 4 bytes.

`<Size>`

Siguiente a eso se codifica la apuesta con toda la información separada por comas

`<Agency>,<Name>,<Surname>,<Document>,<Birthdate>,<Number>`

De esta manera el servidor primero recibe el size de la apuesta que luego usa para recibir la apuesta en sí y hacer un split por comas y obtener la información correctamente.

#### Ejecucion
``` bash
make docker-compose-up
```

### Ejercicio 6
En esta segunda iteración se implementó el envío por batches (varias apuesta a la vez) y además se volvió a utilizar `docker volumes` para usar correctamente los archivos provistos por la cátedra.

Primero cada cliente lee el archivo y se arma en memoria un array de las apuestas que tiene para enviar. Luego se calculan y se envían `4 bytes` al servidor con la información de cuántos batches se van a enviar teniendo en cuenta cuánta apuestas son y el máximo que se puede agrupar por batch.

`<Amount of batches>`

Luego se envían en `4 bytes` el tamaño que tiene el batch y el batch en sí mismo. Para separar las apuestas entre sí se usa el caracter salto de línea (`\n`)

`<Size of batch 1><Batch 1> ... <Size of batch N><Batch N>`

Dentro del batch:

`<Apuesta 1>\n<Apuesta 2>\n...\n<Apuesta N>`

Las apuestas al igual que en el punto anterior están separadas comas

`<Agency>,<Name>,<Surname>,<Document>,<Birthdate>,<Number>`

#### Ejecucion
Igual que el ejercicio anterior
``` bash
make docker-compose-up
```

### Ejercicio 7
En este ejercicio se implementa por primera vez la escritura en el servidor, ya que es necesario para comunicar a las agencias cuáles fueron los ganadores de sus apuestas.

Por parte del cliente, se envían dos tipos de mensajes:

1. Batches de apuestas, que contienen las apuestas realizadas.
2. Mensaje de finalización, que indica que se han enviado todas las apuestas.

Para diferenciar estos mensajes, se introduce un nuevo campo en el protocolo: un flag de 1 byte, que puede tomar los valores `BET_FLAG` o `END_FLAG`, permitiendo así identificar el tipo de mensaje enviado por el cliente.

Adicionalmente, al establecer la conexión con el servidor, el cliente envía un primer mensaje de 1 byte con su identificador de agencia. Esta información permitirá al servidor asociar las apuestas con la agencia correspondiente y, posteriormente, enviarle los resultados.

#### Cliente
En la conexión: `<Identificador de agencia>`

Enviando los batches de apuestas: `<BET_FLAG>` y luego sigue el mismo protocolo que en el ejercicio anterior

Envío de fin de apuestas: `<END_FLAG>` 

#### Servidor
Ahora el servidor verifica qué flag se le envía para saber si almacenar las apuestas al igual que antes o bien si recibe un END_FLAG empezar con el sorteo. Se usan las funciones provistas de `load_bets()` y `has_won()` para armar un diccionario que tiene como key la agencia y como valor un array de documentos de ganadores el cual es enviado a las respectivas agencias con el siguiente formato.

`<Cantidad de ganadores><Documento 1>...<Documento N>`

Tanto la cantidad como los documentos se envían en 4 bytes cada uno.

#### Ejecucion
Igual que el ejercicio anterior
``` bash
make docker-compose-up
```

### Ejercicio 8
En este último ejercicio, solo se modifica el servidor para hacerlo concurrente utilizando la librería `multiprocessing`.

- Al aceptar una nueva conexión, se crea un proceso con `Process`, estableciendo como objetivo la función handle_client_connection que es la encargada de recibir todas las apuestas de la agencia. Una vez que este último recibe la información de los ganadores, se realiza un join del proceso.

- Se utiliza un `Lock` para garantizar un acceso seguro a las funciones load_bets() y store_bets(), ya que estos recursos son compartidos por todos los procesos.

- Para gestionar la terminación del servidor, se emplea un `Value`. Para que en caso de recibir una señal SIGTERM, el proceso debe finalizar, por lo que _server_running se establece en False de manera atómica.

- Finalmente, para la sincronización, se usa una `Barrier` que detiene los procesos justo cuando reciben el END_FLAG, permitiendo procesar a los ganadores con toda la información disponible.

#### Ejecucion
Igual que el ejercicio anterior
``` bash
make docker-compose-up
```