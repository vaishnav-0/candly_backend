## migrations
- create
    migrate create -ext sql -dir db/migrations -seq create_users_table

## Generate key pair for jwt
- openssl genpkey -algorithm ed25519 -outform PEM -out test25519.pem
- openssl pkey -in test25519.pem -pubout -out test25519_pub.pem