# Overview
 
 This repo is a sample of an operator used to trigger watcher of state of a kubernetes cluster applications, to help to simplify installation flows from [kubefirst](https://github.com/kubefirst/kubefirst). 

 This repo is also a guide on how to create from end-to-end a controller for kubernetes CRDs. 


# Tools

In order to simplify in a single place all the tools needed to create a controller ready to use, this repo has a Dockerfile that wraps all needed tools for you. 

To use it is just run: 
```bash 
#Build Container
docker build . -t kubebuilder

#Use Container
docker run --rm -it \
    --name gitlab-bash \
    -v $HOME/.aws:/root/.aws \
    -w /go/src \
    -v operator-sdk:/go/pkg \
    -v $(pwd):/go/src \
    -v /var/run/docker.sock:/var/run/docker.sock  \
    --privileged \
    kubebuilder
```
This container has `--rm` on the run flags, so be aware that everything  out of `/go/src` will removed once you exit the container and you can't recover it. Feel free to map more folders to make easier for your to work with it. 

# CRD 
##  Starting your CRD

This watcher-operators was created 

```bash 
mkdir watcher

cd watcher/

kubebuilder init --domain kubefirst.io --license apache2 --owner "K-rays" --repo github.com/k1tests/basic-controller

kubebuilder create api --group k1 --version v1beta1 --kind Watcher --controller --resource

make manifests

make install
```


## Working

To run your controller run: 

```bash 
cd watcher/

make run

kubectl get watcher

kubectl get crd 
```


