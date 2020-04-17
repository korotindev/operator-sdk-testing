How to run on minikube

```
minikube start --driver=parallels
export OPERATOR_NAME=memcached-operator
operator-sdk run --local --namespace=default
```