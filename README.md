# wock
wock is a simple tool for mocking a host with local files

```shell
$ wock install
Successfully installed/verified local CA
$ wock nytimes.com html
mocking host 'nytimes.com' with files from /c/users/exampleuser/documents/html
```

<p align="center"><img width="440" alt="image" src="https://github.com/cpendery/wock/assets/35637443/faa05394-27ce-4610-981d-0bd3ce574a42"></p>

wock simplifies managing certificates, adjusting host DNS resolution, serving local build files 
all through one unified interface. wock also provides its locally-trusted certificates using 
[mkcert](https://github.com/FiloSottile/mkcert)'s local CA in the system root store.
