class Encoder:
    def __init__(self):
        self._buf = bytearray()
    
    def write_uint8(self, v):
        self._buf += v.to_bytes(1, byteorder='big')

    def write_uint16(self, v):
        self._buf += v.to_bytes(2, byteorder='big')

    def write_uint32(self, v):
        self._buf += v.to_bytes(4, byteorder='big')

    def write_string(self, s):
        encoded = s.encode('utf-8')
        self.write_uint16(len(encoded))
        self._buf += encoded

    def write_fixed_string(self, s):
        self._buf += s.encode('utf-8')
    
    def bytes(self):
        return bytes(self._buf)