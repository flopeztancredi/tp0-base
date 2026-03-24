import socket
import logging
import signal

from common.utils import has_won, load_bets, store_bets
from protocol.protocol import receive_incoming, receive_winners_request, send_winners_response, send_ack

NUM_AGENCIES = 5


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

        # Register signal handler for graceful shutdown
        signal.signal(signal.SIGTERM, self.__handle_sigterm)
    
    def __handle_sigterm(self, signum, frame):
        logging.info('action: graceful_shutdown | result: in_progress')
        self._running = False
        self._server_socket.close()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        self.__receive_bets_phase()
        if not self._running:
            logging.info('action: graceful_shutdown | result: success')
            return

        self.__run_lottery()
        self.__respond_winners_phase()
        logging.info('action: graceful_shutdown | result: success')
    
    def __receive_bets_phase(self):
        done_count = 0
        while self._running and done_count < NUM_AGENCIES:
            try:
                client_sock = self.__accept_new_connection()
            except OSError:
                break
            if self.__handle_incoming(client_sock):
                done_count += 1

    def __handle_incoming(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            kind, data = receive_incoming(client_sock)
            if kind == "batch":
                store_bets(data)
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(data)}")
                send_ack(client_sock, success=True)
                return False
            elif kind == "done":
                logging.info(f"action: agencia_finalizada | result: success | id: {data}")
                return True
        except ValueError as e:
            logging.error(f"action: apuesta_recibida | result: fail | error: {e}")
            send_ack(client_sock, success=False)
            return False
        except (OSError, ConnectionError) as e:
            logging.error(f"action: apuesta_recibida | result: fail | error: {e}")
            return False
        finally:
            client_sock.close()
    
    def __run_lottery(self):
        self._winners_by_agency = {}
        for bet in load_bets():
            if has_won(bet):
                self._winners_by_agency.setdefault(bet.agency, []).append(int(bet.document))
        logging.info('action: sorteo | result: success')
    
    def __respond_winners_phase(self):
        for _ in range(NUM_AGENCIES):
            if not self._running:
                break
            try:
                client_sock = self.__accept_new_connection()
            except OSError:
                break
            self.__handle_winners_request(client_sock)
        
    def __handle_winners_request(self, client_sock):
        try:
            agency = receive_winners_request(client_sock)
            winners = self._winners_by_agency.get(agency, [])
            send_winners_response(client_sock, winners)
            logging.info(f"action: consulta_ganadores | result: success | id: {agency} | cant_ganadores: {len(winners)}")
        except (OSError, ValueError, ConnectionError) as e:
            logging.error(f"action: consulta_ganadores | result: fail | error: {e}")
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
