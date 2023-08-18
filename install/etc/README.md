# Infra Settings

## MetalLB

### Install with Helm

```shell
$ helm repo add metallb https://metallb.github.io/metallb
"metallb" has been added to your repositories

$ helm repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "metallb" chart repository
Update Complete. ⎈Happy Helming!⎈

$ helm repo list
NAME   	URL
metallb	https://metallb.github.io/metallb

$ helm install metallb metallb/metallb --create-namespace -n metallb-system 
NAME: metallb
LAST DEPLOYED: Thu Aug 17 17:23:25 2023
NAMESPACE: metallb-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
MetalLB is now running in the cluster.

Now you can configure it via its CRs. Please refer to the metallb official docs
on how to use the CRs.

$ kubectl apply -f metallb-configs.yaml
ipaddresspool.metallb.io/address-pool created
l2advertisement.metallb.io/layer2 created
```

## ArgoCD

### Install with Manifest

```shell
$ kubectl create namespace argocd
$ kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
customresourcedefinition.apiextensions.k8s.io/applications.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/applicationsets.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/appprojects.argoproj.io created
serviceaccount/argocd-application-controller created
serviceaccount/argocd-applicationset-controller created
serviceaccount/argocd-dex-server created
serviceaccount/argocd-notifications-controller created
serviceaccount/argocd-redis created
...

# option
$ kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}'

# get admin password
$ kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo

# delete all resources
$ kubectl delete -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

## Traefik

### Install with Helm

```shell
$ helm repo add traefik https://traefik.github.io/charts
"traefik" has been added to your repositories

$ helm repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "traefik" chart repository
Update Complete. ⎈Happy Helming!⎈

$ helm repo list
NAME   	URL
traefik	https://traefik.github.io/chart

$ helm install traefik traefik/traefik --create-namespace -n traefik
```