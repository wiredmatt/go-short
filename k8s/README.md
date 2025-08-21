# k8s setup

## dev

```sh
kind create cluster # or use minikube or whatever you prefer

docker build -t myapi:latest . # make sure this is run in the root directory of the project.
kind load docker-image myapi:latest # or use minikube or whatever you prefer

kubectl apply -k k8s/overlays/dev
kubectl rollout restart deployment myapi -n myapp

kubectl get pods -n myapp        
kubectl logs deployment/myapi -n myapp

kubectl port-forward svc/myapi 3000:80 -n myapp
```