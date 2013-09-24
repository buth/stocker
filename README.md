stocker
=======

Stock a docker node. ***VERY ALPHA; DO NOT USE.***

## Encrypted etcd Storage

Just a demo for the moment. With etcd running locally, try the following.

```bash
stocker -k > key.txt
stocker -s key.txt -key test -value 'hello world!'
curl -L http://127.0.0.1:4001/v1/keys/test
stocker -s key.txt -key test
```