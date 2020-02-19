# Using a TPM to sign JWTs

Example uses the opaqueSigner interface in the [go-jose.v2](https://pkg.go.dev/gopkg.in/square/go-jose.v2?tab=doc) 
library to create a signed JWT using a key held securely in a TPM.

##  Build and run

Build:

    docker build -t tpm-jose .
    
Run:

    docker run --rm -v "$PWD"/src:/usr/src/tpm-jose -w /usr/src/tpm-jose tpm-jose go run . -v
