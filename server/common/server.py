import socket
import logging
import signal
from common.bet_center import BetCenterListener, BetCenter
from common.utils import store_bets

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = BetCenterListener.bind('', port, listen_backlog)
        self._server_running = True
        
        signal.signal(signal.SIGTERM, self.__server_handle_sigterm)
        signal.signal(signal.SIGINT, self.__server_handle_sigterm)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._server_running:
            client_sock = self.__accept_new_connection()
            if client_sock:
                try:
                    self.__handle_client_connection(client_sock)
                finally:
                    client_sock.close()

    def __server_handle_sigterm(self, _signal, _frame):
        """
        Maneja la señal SIGTERM para cerrar el servidor de forma segura
        """
        logging.info("action: shutdown | result: success | message: SIGTERM received, shutting down server...")
        self._server_running = False
    
    def __handle_client_connection(self, client_sock: BetCenter):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bets = client_sock.recv()
            store_bets(bets)
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
        except OSError as e:
            logging.error("action: apuesta_recibida | result: fail | error: {e}")

    def __accept_new_connection(self) -> BetCenter:
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except socket.timeout:
            return None
        except OSError as e:
            return None
