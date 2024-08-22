# Playground 

This is a little Playground to build a little k8s cluster with some resources to build scanners on top of it.

## Setup of Local cluster using helm

To setup the playground you need to have a running k8s cluster. You can use e.g. minikube, docker desktop or kind for that

```bash
helm install k8s-playground ./k8s-playground
```

## Usage of Service Account 

To use the Service Account that created with the local cluster helm chart you need to use the "localSAConfig" method in main.go

To generate the token for the Service Account you can use the following command:

```bash
kubectl create token {{ .Values.serviceAccount.name }} --duration=24h > ./my_token
```

## Script examples

There are 3 examples provided in the "main.go" to show how to authenticate against the local cluster and how to use the client-go library to interact with an OIDC protected oidc cluster.

