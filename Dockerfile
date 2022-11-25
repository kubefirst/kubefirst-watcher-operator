FROM golang:1.19

RUN apt-get update \
    && apt-get install -y bash curl git  openssh-client make mercurial openrc docker

# curl for docker image
# git, mercurial, docker for Operator SDK
# bash, openssh, make, openrc for QoL

ARG RELEASE_VERSION=v1.23.0
ARG KUBECTL_VERSION=v1.21.3
ARG OS=linux
ARG ARCH=amd64
ARG OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download

RUN apt-get update && \
    apt-get install -y unzip curl jq vim unzip less \
     && rm -rf /var/lib/apt/lists/* 

# Operator SDK says it needs Kubectl, not  yet sure why though
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/$(uname -s | awk '{print tolower($0)}' )/amd64/kubectl \
    && chmod +x ./kubectl \
    && mv ./kubectl /usr/local/bin/kubectl


# AWS cli
RUN curl -LO https://awscli.amazonaws.com/awscli-exe-$(uname -s| awk '{print tolower($0)}')-x86_64.zip && \
    unzip awscli-exe-$(uname -s| awk '{print tolower($0)}')-x86_64.zip && \
    ./aws/install && \
    rm -r aws && \
    rm awscli-exe-$(uname -s| awk '{print tolower($0)}')-x86_64.zip

# AWS EKS cli
RUN curl -LO https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s| awk '{print tolower($0)}')_amd64.tar.gz && \
    tar -xvzf eksctl_$(uname -s| awk '{print tolower($0)}')_amd64.tar.gz -C /usr/local/bin/ && \
    rm eksctl_$(uname -s| awk '{print tolower($0)}')_amd64.tar.gz

# AWS IAM Authenticator tool
RUN curl -LO https://s3.us-west-2.amazonaws.com/amazon-eks/1.21.2/2021-07-05/bin/$(uname -s| awk '{print tolower($0)}')/amd64/aws-iam-authenticator && \
    chmod +x aws-iam-authenticator && \
    mv aws-iam-authenticator /usr/local/bin/    

# Install Kubebuilder 
#https://book.kubebuilder.io/quick-start.html#installation
RUN curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/${OS}/${ARCH} \
    && chmod +x kubebuilder && mv kubebuilder /usr/local/bin/

# From Operator SDK docs
ENV GO111MODULE=on

# Need /sys/fs/cgroup to not be read-only, when using Docker
VOLUME [ "/sys/fs/cgroup", "/go/src" ]
