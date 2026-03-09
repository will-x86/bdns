# Bad DNS 


The idea here is a social based DNS, where friends share credits / queries.

The goal is not to be a full blown DNS, but rather a beautiful proxy to learn some technologies.


Components: ( tood most LMFAO ) 

- dns - Core DNS using go, goal is to do the following:
    - Parse query ( extract user_id via via sni in DoT/ or subdomain in DoH )
    - Do a simple check if user has the ability to check said side ( credits / full block / whatever)
    - Provide smoochy api for the app





### Gen certs
```bash
openssl genrsa -out server.key 2048
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```
