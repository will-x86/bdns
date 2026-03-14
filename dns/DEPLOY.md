## Work in progress


I'll be using coolify to deploy this app.


I'm testing on an ARM hetzner server, 4gb ram.


This DNS needs to bypass any proxy, primarily as we identify users via SNI ( which is in the TLS handshake)



### Certificates

#### locally
( Just hit enter a bunch ! )
```bash
openssl genrsa -out server.key 2048\
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650\
```
#### production
```bash
sudo apt install snapd
sudo snap install certbot --classic
```

Then run :
```bash
sudo certbot certonly \
  --manual \
  --preferred-challenges dns \
  -d "*.dns.domain.com" \
  -d "domain.com"
```

Obviously change this to your domain, and you may not want to do "domain.com", I did as I'll be using it for https too :) 

<details>
<summary>Output</summary>


You will get prompts about your email & terms and conditions, I already accepted them when I messed up the command before, so they won't be present here.
```
user@server:~# sudo certbot certonly \
  --manual \
  --preferred-challenges dns \
  -d "*.dns.domain.com" \
  -d "domain.com"
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Requesting a certificate for *.dns.domain.com and domain.com

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Please deploy a DNS TXT record under the name:

***TXT NAME***
with the following value:

***TXT CONTENT ***

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Press Enter to Continue

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Please deploy a DNS TXT record under the name:

***TXT NAME ***

with the following value:

***TXT CONTENT ***

(This must be set up in addition to the previous challenges; do not remove,
replace, or undo the previous challenge tasks yet. Note that you might be
asked to create multiple distinct TXT records with the same name. This is
permitted by DNS standards.)

Before continuing, verify the TXT record has been deployed. Depending on the DNS
provider, this may take some time, from a few seconds to multiple minutes. You can
check if it has finished deploying with aid of online tools, such as the Google
Admin Toolbox: https://toolbox.googleapps.com/apps/dig/#TXT/_acme-challenge.will-x86.com.
Look for one or more bolded line(s) below the line ';ANSWER'. It should show the
value(s) you've just added.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Press Enter to Continue

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/dns.domain.com/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/dns.domain.com/privkey.pem
This certificate expires on 2026-06-12.
These files will be updated when the certificate renews.

NEXT STEPS:
- This certificate will not be renewed automatically. Autorenewal of --manual certificates requires the use of an authentication hook script (--manual-auth-hook) but one was not provided. To renew this certificate, repeat this same certbot command before the certificate's expiry date.
```


# Expiring!!!
- This cert will *not* be auto-renewed, and certbot no longer sends emails about expiring certificates
</details>


Now we have our cert, lets deploy it :)

As I said before, I'll be using Coolify so the steps for you may be different.


1. Create a new project
2. Select your method of deployment, for *you* this will be "Git based public repository"
3. Select your server
4. Options:
    - Build pack = Docker compose
    - Base directory = /dns
    - Docker compose location = /prod-docker-compose.yml # Notice the .yml, coolify defaults to .yaml

5. Enviroment:
```
KEY_PATH=/cert/dns.domain.com/privkey.pem
VALKEY_ADDR=valkey:6379
CRT_PATH=/cert/dns.domain.com/fullchain.pem
PORT=853
```
