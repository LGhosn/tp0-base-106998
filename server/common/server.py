import socket
import logging
import signal
from common.bet_center import BetCenterListener, BetCenter
from common.utils import BET_FLAG, END_FLAG, has_won, load_bets, store_bets

class Server:
    def __init__(self, port, listen_backlog, cant_clientes):
        # Initialize server socket
        self.cant_clientes = cant_clientes
        self._server_socket = BetCenterListener.bind('', port, listen_backlog)
        self._server_running = True
        
        signal.signal(signal.SIGTERM, self.__server_handle_sigterm)
        signal.signal(signal.SIGINT, self.__server_handle_sigterm)

    def run(self):
        agencies = {}
        while self._server_running and len(agencies) < self.cant_clientes:
            client_sock = self.__accept_new_connection()
            if client_sock:
                    self.__handle_client_connection(client_sock)
                    agencies[client_sock.agency] = client_sock

        # Close server socket
        self._server_socket.close()

        if self._server_running:
            logging.info("action: sorteo | result: success")

            winners = {}
            for bet in load_bets():
                if has_won(bet):
                    if bet.agency not in winners:
                        winners[bet.agency] = []
                    winners[bet.agency].append(int(bet.document))
            
            #envio los ganadores correspondientes a cada agencia
            for agency, socket in agencies.items():
                socket.send_winners(winners.get(agency, []))

        for socket in agencies.values():
            socket.close()


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
            flag = client_sock.recv()
            if flag == END_FLAG:
                logging.info(f"action: fin_apuestas | result: success")
            elif flag == BET_FLAG:
                data = client_sock.recv_bets()
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(data)}")
                store_bets(data)
            else:
                logging.error(f"action: apuesta_recibida | result: fail | error: invalid flag")
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
