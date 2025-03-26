import socket
from common.utils import BET_FLAG, END_FLAG, Bet
import logging

class BetCenterListener:
    def __init__(self, sock: socket.socket):
        self._socket = sock

    @classmethod
    def bind(cls, host: str, port: int, backlog: int = 0) -> 'BetCenterListener':
        skt = cls(socket.socket(socket.AF_INET, socket.SOCK_STREAM))
        skt._socket.bind((host, port))
        skt._socket.listen(backlog)
        return skt

    
    def accept(self) -> tuple[Bet, tuple[str, int]]:
        client_socket, addr = self._socket.accept()
        agency = int.from_bytes(client_socket.recv(1), byteorder="big")
        return BetCenter(client_socket, agency), addr

    
    def close(self) -> None:
        self._socket.close()

class BetCenter:
    def __init__(self, socket: socket.socket, agency: int):
        self._socket = socket
        self.agency = agency

    def __enter__(self) -> 'BetCenter':
        return self
    
    def __exit__(self, exc_type, exc_value, traceback) -> None:
        self._socket.close()

    def connect(self, host: str, port: int) -> None:
        self._socket.connect((host, port))
        return self._socket

    def recv_all(self, size: int) -> bytes:
        data = bytearray()

        while len(data) < size:
            packet = self._socket.recv(size - len(data), socket.MSG_WAITALL)
            if not packet:
                raise ConnectionError("Socket closed unexpectedly")
            data.extend(packet)

        return bytes(data)
    
    def recv_bets(self) -> list[Bet]:
        num_batches = int.from_bytes(self.recv_all(4), byteorder="big") 
        bets = []

        for _ in range(num_batches):
            batch_size = int.from_bytes(self.recv_all(4), byteorder="big")
            actual_batch = self.recv_all(batch_size)
            for bet in actual_batch.split(b'\n'):
                if bet:
                    bets.append(Bet(*bet.decode().split(',')))

        return bets

    def recv(self) -> list[Bet]:
        flag = int.from_bytes(self.recv_all(1), byteorder="big")
        return flag

    def _sendall(self, data: bytes) -> None:
        sent = 0
        while sent < len(data):
            sent += self._socket.send(data[sent:])

    def send_winners(self, winners: list[str]) -> None:
        # primero envio la cantidad de ganadores
        self._sendall(len(winners).to_bytes(4, byteorder="big"))
        for winner in winners:
            winner_bytes = winner.to_bytes(4, byteorder="big")
            self._sendall(winner_bytes)

    def close(self) -> None:
        self._socket.close()
