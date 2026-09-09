# Informe - Gabriel Re (105095)

## Protocolo de comunicación

### Formato del mensaje

Propongo dividir cada mensaje en un header de tamaño fijo y un payload de tamaño variable:

* **Message Type** (1 byte): Identifica el propósito del mensaje.
* **Payload Length** (4 bytes): Cantidad de bytes en el payload.
* **Payload** (variable): Información correspondiente al tipo de mensaje.

De esta manera el receptor primero debe leer los 5 bytes correspondientes al header. Una vez interpretado el largo del payload, se conoce exactamente cuántos bytes se debe recibir para completar el mensaje. **El orden de bytes es big endian para mantener una representación común entre el cliente y el servidor**.

Payload Length representa unicamente el tamaño del payload y no el tamaño total del mensaje. Por lo tanto:

`message_size = header_size + payload_length`

Se definió además un tamaño máximo de payload de 1 MB. Tanto el cliente como el servidor validan el valor indicado en `Payload Length` antes de procesar el contenido del mensaje, evitando mensajes que excedan el tamaño permitido.

La lectura y escritura sobre el socket no asume que una única operación permita transmitir todos los bytes solicitados. Las funciones implementadas en `safe_socket` continúan leyendo o escribiendo hasta completar la cantidad total de bytes correspondiente al header o al payload.

### Tipos de mensajes

En esta primera versión contemplo cuatro tipos:

- **BET** : Formato que se utiliza para enviar una apuesta desde el cliente hacia el servidor.
- **END_BETS** : Formato que se utiliza para informar que la agencia terminó de enviar apuestas.
- **RESULTS** : Formato que se utiliza para transportar desde el servidor los ganadores correspondientes.
- **ERROR** : Formato que se utiliza para informar errores de protocolo cuando la conexión continúa utilizable.
- **ACK**: Confirma al cliente que todas las apuestas pertenecientes al último batch fueron procesadas correctamente.

Decidí utilizar `END_BETS` explícito en lugar de interpretar `BET` con payload vacío con payload vacío como finalización. Así diferencio claramente los mensajes de datos de los mensajes de control.

De manera similar, no se definió un mensaje especial para una respuesta sin ganadores. Un mensaje `RESULTS` con payload de longitud cero representa válidamente que no existen resultados para esa agencia.

A partir de la incorporación del procesamiento por batches, un mensaje `BET` puede transportar una o más apuestas dentro de un mismo payload. De esta manera el tipo de mensaje continúa representando el envío de apuestas, sin necesidad de agregar un tipo de mensaje específico para batches.

El mensaje `ACK` posee payload vacío. Su función es únicamente indicar que el servidor pudo deserializar y almacenar correctamente todas las apuestas pertenecientes al último mensaje `BET` recibido.

### Serialización de apuestas

Cada apuesta se representa dentro del payload mediante sus campos serializados como texto y separados utilizando el carácter `|`.

El formato utilizado es:

`AgencyId|FirstName|LastName|Id|Birthdate|Number`

Como `|` forma parte del protocolo para separar campos, se valida que los campos de texto de una apuesta no contengan dicho carácter. También se reserva el salto de línea para separar distintas apuestas dentro de un mismo batch.

El cliente recibe mediante la variable de entorno `BATCH_SIZE` la cantidad máxima de apuestas que puede agrupar dentro de cada mensaje `BET`.

Cuando un mensaje contiene múltiples apuestas, cada apuesta se serializa individualmente utilizando el mismo formato y posteriormente se separan mediante `\n`.

Por ejemplo:

`bet1\nbet2\nbet3`

De esta forma el protocolo mantiene una misma representación para una apuesta individual y para un conjunto de apuestas. Al recibir el mensaje, el servidor primero separa el payload utilizando el separador de apuestas y luego deserializa individualmente cada una.

### Flujo inicial de comunicación

Por cada apuesta leída del archivo, el cliente construye una `Bet`,la serializa y la envía al servidor mediante un mensaje `BET`.

A partir del procesamiento por batches, se serializan y se transportan dentro del payload de un único mensaje `BET`.

El servidor recibe cada mensaje, deserializa el payload y almacena la apuesta utilizando `Lottery.store_bets`.

Actualmente, al recibir un mensaje `BET`, el servidor deserializa primero todas las apuestas incluidas dentro del batch. Una vez completada correctamente la deserialización, almacena el conjunto utilizando `Lottery.store_bets`.

El servidor solamente responde mediante un mensaje `ACK` después de haber procesado correctamente todas las apuestas pertenecientes al batch. Si ocurre un error durante la deserialización o el almacenamiento, no se considera confirmado dicho batch.

Luego de enviar un mensaje `BET`, el cliente espera recibir el `ACK` correspondiente antes de continuar con el siguiente batch. Esto permite sincronizar el envío de apuestas con el procesamiento del servidor y evita que el cliente considere procesado un batch antes de recibir su confirmación.

Cuando el cliente termina de recorrer el archivo envía un mensaje `END_BETS` con payload vacío. Este mensaje actúa como mecanismo de sincronización e indica al servidor que puede comenzar a calcular los resultados.

El servidor obtiene las apuestas almacenadas mediante `load_bets`, verifica cada una mediante `has_won` y filtra los ganadores correspondientes a la agencia.

Finalmente, los ganadores se serializan y se envían al cliente mediante un mensaje `RESULTS`. Si la agencia no posee ganadores, `RESULTS` se envía con payload vacío.

El cliente deserializa el payload de `RESULTS` y persiste las apuestas ganadoras en `OUTPUT_FILE`.

### Mecanismos de sincronización

La sincronización entre cliente y servidor se realiza mediante los propios mensajes definidos por el protocolo.

El mensaje `ACK` sincroniza el procesamiento de cada batch. El cliente no envía el siguiente conjunto de apuestas hasta recibir la confirmación de que el batch anterior fue procesado correctamente.

El mensaje `END_BETS` sincroniza la finalización de la carga de apuestas de una agencia. El servidor no calcula ni envía los resultados de dicha agencia hasta haber recibido explícitamente este mensaje.

De esta manera, para una conexión el orden esperado de los mensajes es:

`BET -> ACK -> BET -> ACK -> ... -> END_BETS -> RESULTS`

### Manejo de errores

Se distinguen erroes de transporte y errores de protocolo.

Los errores de transporte, como el cierre del socket antes de completar un header o un payload, son detectados por las funciones de `safe_socket` y se propagan a las capas superiores.

Los errores de protocolo incluyen:

- Recepción de un tipo de mensaje desconocido.
- Payload que no puede ser deserializado.
- Tamaño de payload invalido.

También se valida que los mensajes de control que no requieren información adicional, como `END_BETS` y `ACK`, posean un payload vacío. Por otro lado, un mensaje `BET` debe contener obligatoriamente información para poder ser procesado.

En el caso de los batches, si alguna de las apuestas contenidas dentro del mensaje no puede ser deserializada correctamente, el procesamiento del batch se considera fallido y no se envía la confirmación `ACK`.

Mientras la conexión continúe siendo válida, estos errores podrán informarse mediante un mensaje `ERROR`. Caso de que se produzca la pérdida de la conexión no requiere el envío de dicho mensaje.