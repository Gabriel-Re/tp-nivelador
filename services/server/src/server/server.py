import os
import socket
import logger
import threading # https://docs.python.org/es/3.14/library/threading.html

from lottery import Lottery
from protocol.bet_codec import decode_bets, encode_bets
from protocol.connection import receive_message, send_message, send_error
from protocol.message import MessageType

BETS_FILE_NAME = "bets.csv"


class Server:
    def __init__(self, server_host: str, server_port: int, storage_dir: str, agency_quorum_min: int,) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.agency_quorum_min = agency_quorum_min

        # Creo el directorio para las bets de lottery
        os.makedirs(storage_dir, exist_ok=True)

        storage_path = os.path.join(storage_dir,BETS_FILE_NAME)

        # Sigo con un unico Lottery para los threads
        self.lottery = Lottery(storage_path)

        # Se crea abierto el lock (https://docs.python.org/es/3.14/library/threading.html#lock-objects)
        self.lottery_lock = threading.Lock()
        # Condicion para el quorum
        self.quorum_condition = threading.Condition()

        self.finished_agencies = set()
        self.draw_completed = False

    """
    Registra que una agencia termino de enviar bets y espera hasta
    alcanzar la cantidad mínima de agencias para el sorteo
    """
    def _wait_for_quorum(self, agency_id):
        with self.quorum_condition:
            self.finished_agencies.add(agency_id)

            # PAra debugear
            #logger.info("wait-quorum",logger.LogResult.in_progress,"agency-id",agency_id,"finished-agencies",len(self.finished_agencies),"quorum-min",self.agency_quorum_min)


            # Si se completa el quorum habilita el sorteo y despierta a todos los threads
            if (not self.draw_completed and len(self.finished_agencies) >= self.agency_quorum_min):
                self.draw_completed = True
                self.quorum_condition.notify_all()

            # Si todavía no se alcanzo, libero el lock y espero
            while not self.draw_completed:
                self.quorum_condition.wait()

            # Para debugear
            #logger.info("wait-quorum",logger.LogResult.success,"agency-id",agency_id,"finished-agencies",len(self.finished_agencies),"quorum-min",self.agency_quorum_min)

    """
    Atiende los mensajes recibidos de un cliente
    """
    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        agency_id = None


        # El thread que atiende al cliente cierra el socket al terminar la conexion
        with client_socket:
            try:
                logger.info(action, logger.LogResult.in_progress)

                while True:
                    client_message = receive_message(client_socket)

                    if client_message.header.message_type == MessageType.BET:
                        # Deserializo todas las bets 
                        bets = decode_bets(client_message.payload)

                        # La primera bet recibida identifica el id de la agencia
                        if agency_id is None:
                            agency_id = bets[0].agency_id

                        # Para debugear
                        #logger.info("receive-message",logger.LogResult.success,"agency-id",agency_id,"message-type","BET","bets-amount",len(bets))

                        # Hago lock para evitar problemas de concurrencia
                        with self.lottery_lock:
                            self.lottery.store_bets(bets)

                        message_amount += 1

                        # Respondo con ACK si se procesaron correctamente todas las bets
                        send_message(client_socket,MessageType.ACK)

                        # Para debugear
                        #logger.info("send-message",logger.LogResult.success,"agency-id",agency_id,"message-type","ACK")

                        continue

                    if client_message.header.message_type == MessageType.END_BETS:
                        if agency_id is None:
                            raise ValueError("cannot finish bets without an agency id")

                        # Para debug
                        #logger.info("receive-message",logger.LogResult.success,"agency-id",agency_id,"message-type","END_BETS")

                        # Termino de enviar apuestas y cuento
                        self._wait_for_quorum(agency_id)

                        with self.lottery_lock:
                            winners = [
                                bet
                                for bet in self.lottery.load_bets()
                                if bet.agency_id == agency_id
                                and self.lottery.has_won(bet)
                            ]

                        payload = encode_bets(winners)

                        send_message(client_socket, MessageType.RESULTS, payload)

                        # Para debugear
                        # logger.info("send-message",logger.LogResult.success,"agency-id",agency_id,"message-type","RESULTS","winners-amount",len(winners))

                        logger.info(
                            action,
                            logger.LogResult.success,
                            "messages-amount",
                            message_amount,
                        )

                        return

                    if client_message.header.message_type == MessageType.ERROR:
                        client_error = client_message.payload.decode(
                            "utf-8",
                            errors="replace",
                        )

                        logger.error(
                            action,
                            logger.LogResult.fail,
                            "client-error",
                            client_error,
                        )

                        return

                    raise ValueError(
                        f"unexpected message type: "
                        f"{client_message.header.message_type}"
                    )

            except Exception as e:
                logger.error(
                    action,
                    logger.LogResult.fail,
                    "messages-amount",
                    message_amount,
                )

                error_message = str(e) or "server error"

                try:
                    send_error(client_socket,error_message)
                except Exception:
                    logger.error(
                        "send-error",
                        logger.LogResult.fail,
                    )

    """
    Acepta conexiones y crea un thread independiente para cada cliente
    """
    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                # Cada conexión es procesada por un thread independiente y le indico que empiece en handle_client
                client_thread = threading.Thread(target=self._handle_client,args=(client_socket,))

                try:
                    client_thread.start()

                except Exception:
                    client_socket.close()
                    raise
