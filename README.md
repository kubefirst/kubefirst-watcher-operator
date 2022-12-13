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



# Argo-cd 

In case, the cluster has argocd we can wuse this folowing `patch` to make it to understand state of this new CRD: 
```lua
hs = {}
if obj.status ~= nil then
  if obj.status.status ~= nil then
    if obj.status.status == "Satisfied" then
        hs.status = "Healthy"
        hs.message = obj.status.status
        return hs
     end
     if obj.status.status == "Timeout" then
        hs.status = "Degraded"
        hs.message = obj.status.status
        return hs
     end
  end
end
hs.status = "Progressing"
hs.message = "Waiting for Watcher
return hs
```

```bash 
kubectl patch configmap/argocd-cm \
  -n argocd \
  --type merge \
  -p '{"data":{"resource.customizations.health.k1.kubefirst.io_Watcher":"hs = {}\nif obj.status ~= nil then\n  if obj.status.status ~= nil then\n    if obj.status.status == \"Satisfied\" then\n        hs.status = \"Healthy\"\n        hs.message = obj.status.status\n        return hs\n     end\n     if obj.status.status == \"Timeout\" then\n        hs.status = \"Degraded\"\n        hs.message = obj.status.status\n        return hs\n     end\n  end\nend\nhs.status = \"Progressing\"\nhs.message = \"Waiting for Watcher\"\nreturn hs"} }'
```


# How to install the watcher on your argo installation 

```yaml 
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: k1-watcher
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec: 
  project: default
  source:
    repoURL: 'https://kubefirst.github.io/charts'
    targetRevision: 0.4.0
    helm:
      values: |-
        image: 6zar/k1-watcher-contoller:latest
    chart: helm-k1-watcher-operator
  destination:
    server: 'https://kubernetes.default.svc'
    namespace: watcher-system
  syncPolicy:
      automated:
        prune: true
        selfHeal: true
      syncOptions:
        - CreateNamespace=true
      retry:
        limit: 5
        backoff:
          duration: 5s
          maxDuration: 5m0s
          factor: 2
---
# Creates the argo annotation for CRD
apiVersion: batch/v1
kind: Job
metadata:
  name: add-kubewatcher-argocd
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "0"
spec:
  template:
    spec:
      serviceAccountName: argocd-server
      containers:
      - name: c
        image: portainer/kubectl-shell:latest
        command:
        - /bin/sh
        - -c
        - |
          kubectl patch configmap/argocd-cm \
            -n argocd \
            --type merge \
            -p '{"data":{"resource.customizations.health.k1.kubefirst.io_Watcher":"hs = {}\nif obj.status ~= nil then\n  if obj.status.status ~= nil then\n    if obj.status.status == \"Satisfied\" then\n        hs.status = \"Healthy\"\n        hs.message = obj.status.status\n        return hs\n     end\n     if obj.status.status == \"Timeout\" then\n        hs.status = \"Degraded\"\n        hs.message = obj.status.status\n        return hs\n     end\n  end\nend\nhs.status = \"Progressing\"\nhs.message = \"Waiting for Watcher\"\nreturn hs"} }'
          sleep 10
      restartPolicy: Never
  backoffLimit: 1          
```
