# Playground 

This is a little Playground to build a little k8s cluster with some resources to build scanners on top of it.

## Setup 

To setup the playground you need to have a running k8s cluster. You can use e.g. minikube, docker desktop or kind for that

```bash
helm install k8s-playground ./k8s-playground
```

## Script examples

There are 3 examples provided in the "main.go" to show how to authenticate against the local cluster and how to use the client-go library to interact with an OIDC protected oidc cluster.

```bash