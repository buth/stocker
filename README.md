# Stocker

Stocker provides a method for managing environment variables for a process securely. It was designed with [Docker](https://www.docker.io/) containers in mind, but can be used to set configuration information for any given command.

All information encrypted using [AES-256](http://en.wikipedia.org/wiki/Advanced_Encryption_Standard) in [CBC mode](http://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Cipher-block_chaining_.28CBC.29), signed with a [SHA-256](http://en.wikipedia.org/wiki/SHA-2) [HMAC](http://en.wikipedia.org/wiki/Hash-based_message_authentication_code), and saved in a key-value store.

Stocker is designed to solve the secure configuration issue and *not* to be a full-fledged deployment tool for Docker or anything else.

## Usage

The first step is to create and secure a cryptographic key.

```
$ stocker key /etc/stocker/key
```

Once you have a key, you can use it to set a value.

```
$ stocker set -g mycoolapp PW
PW=(invisibly type value)
```

To make that value available to a running container, use the `stocker exec` command in conjunction with the `-e` flag. Any specified environement variable *not* available will be retrieved and set.

```
$ stocker exec -g mycoolapp -e PW docker run -e PW -e PATH ubuntu env
HOME=/
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=d4929c83a8ea
PW=myawesomevalue
```

## Command Refference

### key

```
stocker key FILENAME
```

The `key` command generates a new cryptographic key to be used in conjunction with the `set` and `exec` commands. The only argument is the filepath to use to save said key to disk. Correct permissions (600) will be set for the created file.

### set

```
stocker set [OPTIONS] VARIABLE [VARIABLE...]
  -b="redis": backend to use
  -g="": group to use for storing and retrieving data
  -h=":6379": backend address
  -k="/etc/stocker/key": path to encryption key
  -n="stocker": backend namespace
  -t="tcp": backend connection protocol
```

The `set` command can be used to save new values for one or more environment variables for a given group (`-g`). After specifying said variables as arguments on the command line, you will be prompted to securely input the coresponding values.

### exec

```
stocker exec [OPTIONS] COMMAND [ARG...]
  -E=false: fetch all environment variables
  -b="redis": backend to use
  -e=[]: environment variable to fetch
  -g="": group to use for storing and retrieving data
  -h=":6379": backend address
  -k="/etc/stocker/key": path to encryption key
  -n="stocker": backend namespace
  -t="tcp": backend connection protocol
  -u="": user to execute the command as
```

The `exec` command will fetch and decode all environment variables (`-E`) for a given group (`-g`) and/or any number of individual environment variables and merge them into the current environment when running the specified command.

## Backends

Stocker is designed to work with any key-value store. At present, the only implemented one is [Redis](http://redis.io/), but it should be straight forward to create backend connectors for [etcd](http://coreos.com/using-coreos/etcd/), [Consul](http://www.consul.io/), and so on.

### Redis (default)

[Redis](http://redis.io/) is the default.

## Contributing

The project is making use of [GitHub issues](https://github.com/blog/831-issues-2-0-the-next-generation) to track progress. If you discover a bug or have a feature request, whether or not you intend to write the code yourself, please open a [new issue](https://github.com/buth/stocker/issues).

For those who want to contribute code, we're using the [git-flow](http://nvie.com/posts/a-successful-git-branching-model/) method, so please submit pull requests as feature branches.