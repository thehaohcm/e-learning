Put the Cloudflare Origin Certificate here (never commit these files):

- origin.pem : the certificate
- origin.key : the private key

Cloudflare dashboard -> SSL/TLS -> Origin Server -> Create Certificate (RSA, 15 years),
then set SSL/TLS encryption mode to "Full (strict)".
Without these files the proxy serves HTTP on port 80 only.
