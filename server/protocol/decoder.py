class Decoder:
    def __init__(self, sock):
        self._sock = sock
    
    def read_uint8(self):
        return int.from_bytes(self._recv_exact(1), byteorder='big')
    
    def read_uint16(self):
        return int.from_bytes(self._recv_exact(2), byteorder='big')
    
    def read_uint32(self):
        return int.from_bytes(self._recv_exact(4), byteorder='big')
    
    def read_string(self):
        length = self.read_uint16()
        return self._recv_exact(length).decode('utf-8')
    
    def read_fixed_string(self, length):
        return self._recv_exact(length).decode('utf-8')

    def _recv_exact(self, n):
        buf = bytearray()
        while len(buf) < n:
            chunk = self._sock.recv(n - len(buf))
            if not chunk:
                raise ConnectionError("Connection closed")
            buf += chunk
        return bytes(buf)