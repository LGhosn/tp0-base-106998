
if [ $# -ne 2 ]; then
    echo "Error: cantidad de parametros incorrecta"
    echo "Uso: $0 <archivo-de-salida> <cantidad-de-clientes>"
    exit 1
fi

# Genero el archivo de salida
echo "name: tp0" > $1
echo "services:" >> $1

# servidor
echo "  server:" >> $1
echo "    container_name: server" >> $1
echo "    image: server:latest" >> $1
echo "    entrypoint: python3 /main.py" >> $1
echo "    environment:" >> $1
echo "      - PYTHONUNBUFFERED=1" >> $1
echo "    networks:" >> $1
echo "      - testing_net" >> $1
echo "    volumes:" >> $1
echo -e "      - ./server/config.ini:/config.ini\n" >> $1


# clientes
for i in $(seq 1 $2); do
    echo "  client$i:" >> $1
    echo "    container_name: client$i" >> $1
    echo "    image: client:latest" >> $1
    echo "    entrypoint: /client" >> $1
    echo "    environment:" >> $1
    echo "      - CLI_ID=$i" >> $1
    echo "      - NOMBRE=Lionel Andres" >> $1
    echo "      - APELLIDO=Messi" >> $1
    echo "      - DOCUMENTO=106998" >> $1
    echo "      - FECHA_NACIMIENTO=1999-03-17" >> $1
    echo "      - NUMERO=7574" >> $1
    echo "    networks:" >> $1
    echo "      - testing_net" >> $1
    echo "    depends_on:" >> $1
    echo "      - server" >> $1
    echo "    volumes:" >> $1
    echo -e "      - ./client/config.yaml:/config.yaml\n" >> $1
done


# red
echo "networks:" >> $1
echo "  testing_net:" >> $1
echo "    ipam:" >> $1
echo "      driver: default" >> $1
echo "      config:" >> $1
echo "        - subnet: 172.25.125.0/24" >> $1

echo "archivo generado correctamente."

exit 0
