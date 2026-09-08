Redactar un breve informe en donde se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado y los mecanismos para sincronizar la ejecución concurrente.

# Informe - Gabriel Re (105095)

## Protocolo de comunicación

### Formato inicial del mensaje

Propongo dividir cada mensaje en un header de tamaño fijo y un payload de tamaño variable:

- *Message Type* (1 byte): Identifica el propósito del mensaje.
- *Payload Length* (4 bytes): Cantidad de bytes en el payload.
- *Payload* (variable): Información correspondiente al tipo de mensaje.

De esta manera el receptor primero debe leer los 5 bytes correspondientes al header. Una vez interpretado el largo del payload, se conoce exactamente cuántos bytes se debe recibir para completar el mensaje. **El orden de bytes es big endian para mantener una representación común entre el cliente y el servidor**.

Payload Length representa unicamente el tamaño del payload y no el tamaño total del mensaje. Por lo tanto:

`message_size = header_size + payload_length`

### Tipos de mensajes

En esta primera versión contemplo cuatro tipos:

- *BET* : Formato que se utiliza para enviar una apuesta desde el cliente hacia el servidor.
- *END_BETS* : Formato que se utiliza para informar que la agencia terminó de enviar apuestas.
- *RESULTS* : Formato que se utiliza para transportar desde el servidor los ganadores correspondientes.
- *ERROR* : Formato que se utiliza para informar errores de protocolo cuando la conexión continúa utilizable.

Decidí utilizar `END_BETS` explícito en lugar de interpretar `BET` con payload vacío con payload vacío como finalización. Así diferencio claramente los mensajes de datos de los mensajes de control.

De manera similar, no se definió un mensaje especial para una respuesta sin ganadores. Un mensaje `RESULTS` con payload de longitud cero representa válidamente que no existen resultados para esa agencia.

### Flujo inicial de comunicación

Por cada apuesta leída del archivo, el cliente construye una `Bet`,la serializa y la envía al servidor mediante un mensaje `BET`.

El servidor recibe cada mensaje, deserializa el payload y almacena la apuesta utilizando `Lottery.store_bets`.

Cuando el cliente termina de recorrer el archivo envía un mensaje `END_BETS` con payload vacío. Este mensaje actúa como mecanismo de sincronización e indica al servidor que puede comenzar a calcular los resultados.

El servidor obtiene las apuestas almacenadas mediante `load_bets`, verifica cada una mediante `has_won` y filtra los ganadores correspondientes a la agencia.

Finalmente, los ganadores se serializan y se envían al cliente mediante un mensaje `RESULTS`. Si la agencia no posee ganadores, `RESULTS` se envía con payload vacío.

El cliente deserializa el payload de `RESULTS` y persiste las apuestas ganadoras en `OUTPUT_FILE`.

### Manejo de errores

Se distinguen erroes de transporte y errores de protocolo.

Los errores de transporte, como el cierre del socket antes de completar un header o un payload, son detectados por las funciones de `safe_socket` y se propagan a las capas superiores.

Los errores de protocolo incluyen:

- Recepción de un tipo de mensaje desconocido.
- Payload que no puede ser deserializado.
- Tamaño de payload invalido.

Mientras la conexión continúe siendo válida, estos errores podrán informarse mediante un mensaje `ERROR`. Caso de que se produzca la pérdida de la conexión no requiere el envío de dicho mensaje.