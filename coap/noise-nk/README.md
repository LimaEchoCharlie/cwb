# COAP client-server with Noise Protocol NK example

Example uses the NK Noise protocol for the message encryption.

* **N**o static key for client.
* The server's static key is **K**nown to the client.

This handshake consists of a single request and response. 
Since the client has pre-knowledge of the server's static key, we can use zero round trip encryption
and encrypt the client request in the first handshake payload.
