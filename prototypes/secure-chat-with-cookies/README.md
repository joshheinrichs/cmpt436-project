# secure-chat

This is intended to be a simple prototype for testing https and wss. None of the chat system has been really been implemented yet, but it will model the redigo-chat prototype when it's done.

### Setup

Setup is a bit more involved for this prototype. Make sure you have [Bower](http://bower.io/) and [Go](https://golang.org/) installed, and then run:

`bower install`

in the `app/` directory to get all of the website dependencies. After that, you will need to create a certificate and key (named `cert.crt` and `key.key`) for https and wss connection, and place them in the `server/` directory. I'd recommend following [this tutorial](https://www.digitalocean.com/community/tutorials/how-to-create-an-ssl-certificate-on-nginx-for-ubuntu-14-04) to generate self-signed certificates.

After that, run

`go run main.go`

in the `server/` directory, and the site will be accessible via [http://localhost:8000/](http://localhost:8000/)
