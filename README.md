# How to run on minikube

## Run locally
```shell script
minikube start --driver=parallels
make code-gen
make apply
make test-e2e
```
## Run e2e tests
```shell script
minikube start --driver=parallels
make code-gen
make apply
make run
```

## Cleanup
```shell script
minikube delete
```

## Create Application
```shell script
kubectl apply deploy/crds/app.korotin.dev_v1alpha1_application_cr.yaml
kubectl describe applications                                                                                                                                                                  soxat.local: Fri Apr 17 16:28:11 2020
```
You will see:
```
Name:         example-application
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"app.korotin.dev/v1alpha1","kind":"Application","metadata":{"annotations":{},"name":"example-application","namespace":"defau...
API Version:  app.korotin.dev/v1alpha1
Kind:         Application
Metadata:
  Creation Timestamp:  2020-04-17T11:43:52Z
  Generation:          3
  Managed Fields:
    API Version:  app.korotin.dev/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:containers:
        f:replicas:
    Manager:      kubectl
    Operation:    Update
    Time:         2020-04-17T12:21:10Z
    API Version:  app.korotin.dev/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:pods:
        f:replicas:
    Manager:         operator-sdk-testing-local
    Operation:       Update
    Time:            2020-04-17T12:21:51Z
  Resource Version:  15917
  Self Link:         /apis/app.korotin.dev/v1alpha1/namespaces/default/applications/example-application
  UID:               a1552045-5517-4d3a-a47a-a0992ac39c06
Spec:
  Containers:
    Cpu Limit:     text
    Image:         nginx:latest
    Memory Limit:  text
    Name:          nginx
    Ports:
      Container Port:  80
      Host Port:       80
      Name:            default
  Replicas:            3
Status:
  Pods:
    example-application-847bc76779-7ljfg
    example-application-847bc76779-d55rh
    example-application-847bc76779-wk7ls
  Replicas:  3
Events:      <none>
```