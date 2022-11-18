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

kubebuilder init --domain kubefirst.io --license apache2 --owner "K-rays" --repo github.com/k1tests/basic-controller

kubebuilder create api --group k1 --version v1beta1 --kind Watcher --controller --resource

make manifests

make install

make run

kubectl get watcher

kubectl get crd 
```

# Sample Watcher
```yaml 
apiVersion: k1.kubefirst.io/v1beta1
kind: Watcher
metadata:
  name: watcher-sample-01
spec:
    size: 1
```
