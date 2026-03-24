from common.utils import Bet
from protocol.decoder import Decoder
from protocol.encoder import Encoder

MSG_TYPE_BET = 0x01
MSG_TYPE_ACK = 0x02
MSG_TYPE_BATCH = 0x03
ACK_SUCCESS  = 0x00
ACK_FAILURE  = 0x01

def receive_batch(sock):
    dec = Decoder(sock)
    msg_type = dec.read_uint8()
    if msg_type != MSG_TYPE_BATCH:
        raise ValueError(f"Expected batch message, got {msg_type}")
    
    count = dec.read_uint16()
    bets = []
    for _ in range(count):
        agency = str(dec.read_uint32())
        first_name = dec.read_string()
        last_name = dec.read_string()
        document = str(dec.read_uint32())
        birthdate = dec.read_fixed_string(10)
        number = str(dec.read_uint32())
        bets.append(Bet(agency, first_name, last_name, document, birthdate, number))

    return bets

def send_ack(sock, success):
    enc = Encoder()
    enc.write_uint8(MSG_TYPE_ACK)
    enc.write_uint8(ACK_SUCCESS if success else ACK_FAILURE)
    sock.sendall(enc.bytes())