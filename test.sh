printf "-----mTLS with curl key 1 \n\n\n\n-----\n"

curl -k -i \
--cert keys/clients/cert-1.pem --key keys/clients/key-1.pem \
--resolve myapp.example.com:6196:127.0.0.1 \
https://myapp.example.com:6196/


printf "------mTLS with curl key 2 \n\n\n\n-----\n"


curl -k -i \
--cert keys/clients/cert-2.pem --key keys/clients/key-2.pem \
--resolve myapp.example.com:6196:127.0.0.1 \
https://myapp.example.com:6196/


printf "------mTLS with curl invalid key \n\n\n\n-----\n"


curl -k -i  --cert keys/clients/invalid-cert.pem --key keys/clients/invalid-key.pem --resolve myapp.example.com:6196:127.0.0.1 https://myapp.example.com:6196/

