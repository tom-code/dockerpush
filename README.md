# dockerpush
minimal code to push single layer image to docker image repository (without docker)

- put your data in tar file then run for example:
```
go run push.go defs.go -image iname -tag v1 -repo http://192.168.1.51:5000 -tar image.tar
```
- for two layer image:
```
go run push.go defs.go -image iname -tag v1 -repo http://192.168.1.51:5000 -tar image.tar -tar2 alpine.tar
```
- then test:
```docker pull localhost:5000/iname:v1```
- works with default setup of private repo https://docs.docker.com/registry/deploying/
- not suitable for large images - full image is loaded in memory for now
