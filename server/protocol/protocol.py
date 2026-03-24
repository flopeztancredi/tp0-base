from common.utils import Bet
from protocol.decoder import Decoder
from protocol.encoder import Encoder

MSG_TYPE_ACK              = 0x02
MSG_TYPE_BATCH            = 0x03
MSG_TYPE_DONE             = 0x04
MSG_TYPE_WINNERS_REQUEST  = 0x05
MSG_TYPE_WINNERS_RESPONSE = 0x06
ACK_SUCCESS  = 0x00
ACK_FAILURE  = 0x01

def receive_incoming(sock):
    dec = Decoder(sock)
    msg_type = dec.read_uint8()
    if msg_type == MSG_TYPE_BATCH:
        return "batch", _receive_batch_body(dec)
    elif msg_type == MSG_TYPE_DONE:
        return "done", _receive_done_body(dec)
    else:
        raise ValueError(f"Unknown message type: {msg_type}")

def _receive_batch_body(dec):
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

def _receive_done_body(dec):
    return dec.read_uint32()

def receive_winners_request(sock):
    dec = Decoder(sock)
    msg_type = dec.read_uint8()
    if msg_type != MSG_TYPE_WINNERS_REQUEST:
        raise ValueError(f"Expected winners request message, got {msg_type}")
    return dec.read_uint32()

def send_winners_response(sock, winners):
    enc = Encoder()
    enc.write_uint8(MSG_TYPE_WINNERS_RESPONSE)
    enc.write_uint16(len(winners))
    for winner in winners:
        enc.write_uint32(winner)
    sock.sendall(enc.bytes())

def send_ack(sock, success):
    enc = Encoder()
    enc.write_uint8(MSG_TYPE_ACK)
    enc.write_uint8(ACK_SUCCESS if success else ACK_FAILURE)
    sock.sendall(enc.bytes())