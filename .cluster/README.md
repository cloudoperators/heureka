How to make Concourse to access the heureka-dev cluster
=======================================================

1. Create a service account in the cluster:
```bash
kubectl apply -f serviceaccount.yaml
```

2. Create a cluster role binding for the service account:
```bash
kubectl apply -f clusterrolebinding.yaml
```

3. Create a token for the service account:
```bash
kubectl apply -f token.yaml
```

4. Verify that the token is created:
```bash
kubectl get secrets -n kube-system
```

5. Get the token:
```bash
kubectl get secret ci-token -n kube-system -o jsonpath="{.data.token}" | base64 --decode
```

6. Add the token to the vault `secrets/` path. Concourse will look in the following paths:
   - `concourse/common/`
   - `concourse/teams/$TEAM/`
   - `concourse/pipelines/$TEAM/$PIPELINE/`
