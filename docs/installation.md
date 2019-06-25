# Installation guide

## Building Docker image
Go to the root directory of the Network Resources Injector and build image:
```
cd $GOPATH/src/github.com/intel/network-resources-injector
make image
```

## Deploying webhook application
Create ssl certificate file which is used for admission controller:
```
./scripts/webhook-create-signed-cert.sh
```

> Note: If you want to use non-self-signed certificate, you just create secret resource with following command:

```
kubectl create secret generic network-resources-injector-secret \
        --from-file=key.pem=<your server-key.pem> \
        --from-file=cert.pem=<your server-cert.pem> \
        -n kube-system
```

Next step create the following resources required to run webhook:
* mutating webhook configuration
* service to expose webhook deployment to the API server

```
kubectl apply -f deployments/service.yaml
cat deployments/webhook.yaml | ./scripts/webhook-patch-ca-bundle.sh | kubectl create -f -
```

Start the actual webhook server application container.

Execute command:
```
kubectl apply -f deployments/server.yaml
```

> Note: Verify that Kubernetes controller manager has --cluster-signing-cert-file and --cluster-signing-key-file parameters set to paths to your CA keypair to make sure that Certificates API is enabled in order to generate certificate signed by cluster CA. More details about TLS certificates management in a cluster available [here](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/).*
