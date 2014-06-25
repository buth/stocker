# Stocker

Stocker provides a method for managing environment variables for a process securely. It was designed with [Docker](https://www.docker.io/) containers in mind, but can be used to set configuration information for any given command.

When run as a server, Stocker accepts SSH connections from Stocker clients for both **writers** and **readers**. Authorized public keys are retrived for both users when the server is started. Values are encrypted and decrypted as requested using a seperate private key stored only on the server; this means that client keys can be rotated, added to, revoked, etc. without the need to re-encrypt data in the key/value store backend.

![Stocker](https://s3.amazonaws.com/newsdev-pub/info/stocker/stocker.svg?v=0.5.1)

Stocker is designed to work with any backend, but presently only [Redis](http://redis.io/) has been implemented. All information stored with a given backend is encrypted using [AES-256](http://en.wikipedia.org/wiki/Advanced_Encryption_Standard) in [CBC mode](http://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Cipher-block_chaining_.28CBC.29), signed with a [SHA-512](http://en.wikipedia.org/wiki/SHA-2) [HMAC](http://en.wikipedia.org/wiki/Hash-based_message_authentication_code).

Stocker is designed to solve the secure configuration issue and *not* to be a full-fledged deployment tool for Docker or anything else.


## Command Refference

### key

```
stocker key filename
```

The `key` command generates a new cryptographic key to be used in conjunction with the `server` command. The only argument is the filepath to use to save said key to disk. Correct permissions (600) will be set for the created file.

### set

```
stocker set [options] variable [variable...]
  -E=false: use current environment when possible
  -a=":2022": address of the stocker server
  -g="": group to use for storing and retrieving data
  -i="": path to an SSH private key
```

The `set` command can be used to save new values for one or more environment variables for a given group (`-g`). After specifying said variables as arguments on the command line, you will be prompted to securely input the coresponding values.

### exec

```
stocker exec [options] command [argument...]
  -a=":2022": address of the stocker server
  -g="": group to use for storing and retrieving data
  -i="": path to an SSH private key
  -u="": user to execute the command as
```

The `exec` command will fetch and decode all environment variables (`-E`) for a given group (`-g`) and/or any number of individual environment variables and merge them into the current environment when running the specified command.

### server

```
stocker server [options]
  -a=":2022": address to listen on
  -b="redis": backend to use
  -h=":6379": backend address
  -i="/etc/stocker/id_rsa": path to an ssh private key
  -k="/etc/stocker/key": path to encryption key
  -n="stocker": backend namespace
  -r="": retrieve reader public keys from this URL
  -t="tcp": backend connection protocol
  -w="": retrieve writer public keys from this URL

```

The `server` command will run a new Stocker server process in the foreground.

## Contributing

The project is making use of [GitHub issues](https://github.com/blog/831-issues-2-0-the-next-generation) to track progress. If you discover a bug or have a feature request please open a [new issue](https://github.com/buth/stocker/issues/new), regardless of whether or not you intend to contribute code yourself.

For those who want to contribute code, we're using the [git-flow](http://nvie.com/posts/a-successful-git-branching-model/) method, so please submit pull requests as feature branches.