# basic-controller

# kubebuilder-sandbox
Explore kube builder to create controller


# Running 

```bash 
docker build . -t kubebuilder


docker run --rm -it \
    -w /go/src \
    -v operator-sdk:/go/pkg \
    -v $(pwd):/go/src \
    --privileged \
    kubebuilder



mkdir sample

cd sample/

kubebuilder init --domain my.domain --repo my.domain/guestbook --plugins=go/v4-alpha

kubebuilder create api --group webapp --version v1 --kind Guestbook
```
