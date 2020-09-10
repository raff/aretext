# This file is generated by gen.go.  DO NOT EDIT.

from __future__ import annotations
from dataclasses import dataclass, asdict
import json
import os
from socket import socket
from typing import Dict
from urllib.parse import urlparse


API_VERSION = "17a79c7761a9a7094bbc7f84764f5040"
API_ADDR = os.environ["API_ADDRESS"]
API_KEY = os.environ["API_KEY"]


class Client:
    """A client for sending RPCs to the aretext API."""

    def __init__(self, addr: str, api_key: str):
        self._socket = socket()
        self._api_key = api_key
        parsed_addr = urlparse("tcp://{}".format(addr))
        self._socket.connect((parsed_addr.hostname, parsed_addr.port))

    def disconnect(self):
        self._socket.close()

    def quit(self, msg: EmptyMsg) -> QuitResultMsg:
        """Quit the aretext editor."""
        self._send("quit", asdict(msg))
        return QuitResultMsg(**self._receive())

    def _send(self, endpoint: str, msg: Dict):
        header = {
            "api_version": API_VERSION,
            "api_key": self._api_key,
            "endpoint": endpoint,
        }
        self._send_frame(self._serialize(header))
        self._send_frame(self._serialize(msg))

    def _receive(self) -> Dict:
        header = self._deserialize(self._receive_frame())
        if not header["success"]:
            raise ServerError(header["error"])
        return self._deserialize(self._receive_frame())

    def _send_frame(self, frame_data: str):
        frame_len = len(frame_data).to_bytes(4, byteorder="big")
        self._socket.sendall(frame_len)
        self._socket.sendall(frame_data)

    def _receive_frame(self) -> bytes:
        frame_len = int.from_bytes(self._socket.recv(4), byteorder="big")
        return self._socket.recv(frame_len)

    @staticmethod
    def _serialize(msg: Dict) -> bytes:
        return json.dumps(msg).encode("utf8")

    @staticmethod
    def _deserialize(data: bytes) -> Dict:
        return json.loads(data.decode("utf8"))


DEFAULT_CLIENT = Client(API_ADDR, API_KEY)


@dataclass
class ServerError(Exception):
    """The server responded with an error."""

    msg: str


@dataclass
class EmptyMsg:
    """
    A message with no fields.

    """

    pass


@dataclass
class QuitResultMsg:
    """
    A message describing the result of a quit request.

    Fields:

            accepted (bool): Whether the quit request was accepted.
            reject_reason (string): The reason the quit request was rejected.
    """

    accepted: bool = True

    reject_reason: str = ""
