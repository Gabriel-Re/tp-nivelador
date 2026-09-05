Redactar un breve informe en donde se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado y los mecanismos para sincronizar la ejecución concurrente.

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

**TODO**

### Manejo de errores

Se distinguen erroes de transporte y errores de protocolo.

Los errores de transporte, como el cierre del socket antes de completar un header o un payload, son detectados por las funciones de `safe_socket` y se propagan a las capas superiores.

Los errores de protocolo incluyen:

- Recepción de un tipo de mensaje desconocido.
- Payload que no puede ser deserializado.
- Tamaño de payload invalido.

Mientras la conexión continúe siendo válida, estos errores podrán informarse mediante un mensaje `ERROR`. Caso de que se produzca la pérdida de la conexión no requiere el envío de dicho mensaje.