import socket
from common.utils import Bet
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
        return BetCenter(client_socket), addr

    
    def close(self) -> None:
        self._socket.close()

class BetCenter:
    def __init__(self, socket: socket.socket):
        self._socket = socket

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
            packet = self._socket.recv(size - len(data))
            if not packet:
                break
            data.extend(packet)

        return bytes(data)
    
    def recv(self) -> list[Bet]:
        all_batch_size = int.from_bytes(self.recv_all(4), byteorder="big") 
        bets = []

        for _ in range(all_batch_size):
            batch_size = int.from_bytes(self.recv_all(4), byteorder="big")
            actual_batch = self.recv_all(batch_size)
            for bet in actual_batch.split(b'\0'):
                if bet:
                    bets.append(Bet(*bet.decode().split(',')))

        return bets


    def close(self) -> None:
        self._socket.close()
