#!/bin/bash

if [ $# -ne 0 ]; then
    echo "Error: cantidad de parámetros incorrecta"
    echo "Uso: $0"
    exit 1
fi

SERVER_IP=$(grep SERVER_IP server/config.ini | cut -d ' ' -f3)
SERVER_PORT=$(grep SERVER_PORT server/config.ini | cut -d ' ' -f3)

if [[ -z "$SERVER_IP" || -z "$SERVER_PORT" ]]; then
    echo "Error: No se pudo obtener la IP o el puerto del servidor."
    exit 1
fi

# Mensaje de prueba
MSG="DURAZNO EN ALMIBAR"

RESPONSE=$(docker run --rm --network=tp0_testing_net --entrypoint sh subfuzion/netcat -c "echo \"$MSG\" | nc -w 20 $SERVER_IP $SERVER_PORT")

if [ "$RESPONSE" = "$MSG" ]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi
