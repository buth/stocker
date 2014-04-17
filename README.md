# Stocker

Stocker provides a method for storing secure configuration information and passing it on to running [Docker](https://www.docker.io/) containers. All information is encrypted using [AES-256](http://en.wikipedia.org/wiki/Advanced_Encryption_Standard) in [CBC mode](http://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Cipher-block_chaining_.28CBC.29) and signed with a [SHA-256](http://en.wikipedia.org/wiki/SHA-2) [HMAC](http://en.wikipedia.org/wiki/Hash-based_message_authentication_code).

Stocker is designed to solve the secure configuration issue and *not* to be a full-fledged Docker deployment tool.

## Usage

The first step is to create and secure a cryptographic key.

```
$ stocker key > key.txt
$ chmod 600 key.txt
```

Once you have a key, you can use it to set a value.

```
$ stocker -secret key.txt set mycoolapp PASSWORD
PASSWORD=(invisibly type value)
```

To make that value available to a running container, use `stocker run` instead of `docker run` and pass in unset environement variables.

```
$ stocker -secret key.txt run mycoolapp -e PASSWORD ubuntu /usr/bin/env
HOME=/
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=d4929c83a8ea
PASSWORD=myawesomevalue
```

## Backends

Stocker is designed to work with any key value store. At present, the only implemented one is [Redis](http://redis.io/), but it should be straight forward to create backend connectors for [etcd](http://coreos.com/using-coreos/etcd/), [Consul](http://www.consul.io/), and so on.

If you are interested in contributing please do!

## Auditing

The cryptographic portions of the code are commented and hopefully easily decipherable. If you want to help expand test coverage or find an issue please pass it on!

