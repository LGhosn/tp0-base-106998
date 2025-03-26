import socket
import logging
import signal
import multiprocessing
from time import sleep
from common.bet_center import BetCenterListener, BetCenter
from common.utils import BET_FLAG, END_FLAG, has_won, load_bets, store_bets

class Server:
    def __init__(self, port, listen_backlog, cant_clientes):
        self.cant_clientes = cant_clientes
        self._server_socket = BetCenterListener.bind('', port, listen_backlog)
        self._server_running = multiprocessing.Value('b', True)
        self.agencies = {}
        self.end_flags = multiprocessing.Barrier(cant_clientes)
        self.bets_lock = multiprocessing.Lock()

        signal.signal(signal.SIGTERM, self.__server_handle_sigterm)
        signal.signal(signal.SIGINT, self.__server_handle_sigterm)

    def run(self):
        while self._server_running.value and len(self.agencies) < self.cant_clientes:
            client_sock = self.__accept_new_connection()
            if client_sock:
                client_process = multiprocessing.Process(
                    target=self.__handle_client_connection,
                    args=(client_sock, self.end_flags, self.bets_lock)
                )
                self.agencies[client_sock.agency] = (client_process, client_sock)
                client_process.start()

        
        for p, s in self.agencies.values():
            p.join()
            s.close()
        
        self._server_socket.close()
        sleep(5)

    def __process_results(self, client_sock):
        logging.info("action: sorteo | result: success")

        winners = []
        for bet in load_bets():
            if bet.agency == client_sock.agency and has_won(bet):
                winners.append(int(bet.document))
        # Enviar ganadores a cada cliente
        client_sock.send_winners(winners)

    def __server_handle_sigterm(self, _signal, _frame):
        logging.info("action: shutdown | result: success | message: SIGTERM received, shutting down server...")
        self._server_running.value = False  

    def __handle_client_connection(self, client_sock: BetCenter, end_flags, bets_lock):
        """
        Maneja la comunicación con un cliente en un proceso separado.
        """
        while self._server_running:
            try:
                flag = client_sock.recv()
                if flag == BET_FLAG:
                    data = client_sock.recv_bets()
                    logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(data)}")
                    bets_lock.acquire()
                    store_bets(data)
                    bets_lock.release()

                elif flag == END_FLAG:
                    logging.info(f"action: fin_apuestas | result: success")
                    end_flags.wait(timeout=30)
                    bets_lock.acquire()
                    self.__process_results(client_sock)
                    bets_lock.release()
                    break
                else:
                    raise OSError(f"flag desconocido {flag} from {client_sock.agency}")

            except OSError as e:
                logging.error(f"action: apuesta_recibida | result: fail | error: {e}")

    def __accept_new_connection(self) -> BetCenter:
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except (socket.timeout, OSError):
            return None
