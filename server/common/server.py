import socket
import logging
import signal
import threading

from common.utils import has_won, load_bets, store_bets
from protocol.protocol import receive_incoming, receive_winners_request, send_winners_response, send_ack


class Server:
    def __init__(self, port, listen_backlog, num_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

        # Initialize structures for multithreading
        self._bets_lock = threading.Lock()
        self._barrier = threading.Barrier(num_agencies, action=self.__run_lottery)
        self._winners_by_agency = {}

        # Register signal handler for graceful shutdown
        signal.signal(signal.SIGTERM, self.__handle_sigterm)
    
    def __handle_sigterm(self, signum, frame):
        logging.info('action: graceful_shutdown | result: in_progress')
        self._running = False
        self._server_socket.close()

    def run(self):
        """
        Server loop

        Accepts new connections indefinitely and spawns a thread per client.
        Terminates on SIGTERM
        """

        while self._running:
            try:
                client_sock = self.__accept_new_connection()
            except OSError:
                break
            t = threading.Thread(target=self.__handle_client, args=(client_sock,), daemon=True)
            t.start()

        logging.info("action: graceful_shutdown | result: success")
    
    def __handle_client(self, client_sock):
        try:
            kind, data = receive_incoming(client_sock)
            if kind == "batch":
                with self._bets_lock:
                    store_bets(data)
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(data)}")
                send_ack(client_sock, success=True)
                client_sock.close()
            elif kind == "done":
                logging.info(f"action: agencia_finalizada | result: success | id: {data}")
                self._barrier.wait()
                self.__handle_winners_request(client_sock)
        except ValueError as e:
            logging.error(f"action: apuesta_recibida | result: fail | error: {e}")
            send_ack(client_sock, success=False)
            client_sock.close()
        except (OSError, ConnectionError) as e:
            logging.error(f"action: apuesta_recibida | result: fail | error: {e}")
            client_sock.close()
        except threading.BrokenBarrierError as e:
            logging.error(f"action: sorteo | result: fail | error: {e}")
            client_sock.close()

    def __run_lottery(self):
        self._winners_by_agency = {}
        for bet in load_bets():
            if has_won(bet):
                self._winners_by_agency.setdefault(bet.agency, []).append(int(bet.document))
        logging.info('action: sorteo | result: success')
        
    def __handle_winners_request(self, client_sock):
        try:
            agency = receive_winners_request(client_sock)
            winners = self._winners_by_agency.get(agency, [])
            send_winners_response(client_sock, winners)
            logging.info(f"action: winners_sent | result: success | id: {agency} | cant_ganadores: {len(winners)}")
        except (OSError, ValueError, ConnectionError) as e:
            logging.error(f"action: winners_sent | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
