# Stocker

Stocker provides a method for managing secure environment variables for a process. It was designed with [Docker](https://www.docker.io/) containers in mind, but can be used to set configuration information for any given command.

All information encrypted using [AES-256](http://en.wikipedia.org/wiki/Advanced_Encryption_Standard) in [CBC mode](http://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Cipher-block_chaining_.28CBC.29), signed with a [SHA-256](http://en.wikipedia.org/wiki/SHA-2) [HMAC](http://en.wikipedia.org/wiki/Hash-based_message_authentication_code), and stored in a settable backed key-value store.

Stocker is designed to solve the secure configuration issue and *not* to be a full-fledged deployment tool for Docker or anything else.

## Usage

The first step is to create and secure a cryptographic key.

```
$ stocker key > key.txt
$ chmod 600 key.txt
```

Once you have a key, you can use it to set a value.

```
$ stocker -k key.txt -e PW set mycoolapp
PW=(invisibly type value)
```

To make that value available to a running container, use the `stocker exec` command in conjunction with the `-e` flag. Any specified environement variable *not* available will be retrieved and set.

```
$ stocker -k key.txt -e PW -e PATH exec mycoolapp docker run -e PW ubuntu /usr/bin/env
HOME=/
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=d4929c83a8ea
PW=myawesomevalue
```

## Backends

Stocker is designed to work with any key value store. At present, the only implemented one is [Redis](http://redis.io/), but it should be straight forward to create backend connectors for [etcd](http://coreos.com/using-coreos/etcd/), [Consul](http://www.consul.io/), and so on.

**If you are interested in contributing please do!**

## Auditing

The cryptographic portions of the code are commented and hopefully easily decipherable. 

**If you find an issue please pass it on!**

## Roadmap
* Switch to using the `syscall` package instead of `os/exec`.
* Add ability to set user and group for the sub-process.

## Contributing

The project is using the [git-flow](http://nvie.com/posts/a-successful-git-branching-model/) method, so please submit pull requests as feature branches.