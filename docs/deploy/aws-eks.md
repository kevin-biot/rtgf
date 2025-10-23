# RTGF Registry Deployment on AWS EKS (Draft)

This guide outlines the steps to deploy the RTGF Authoritative Registry on AWS Elastic Kubernetes Service (EKS).

## Prerequisites
- EKS cluster (Kubernetes v1.28 or newer)
- `kubectl` and Helm v3
- AWS Load Balancer Controller (ALB Ingress)
- `cert-manager` (for TLS) or ACM integration
- Optional: `external-dns` to manage Route53 records
- AWS KMS asymmetric key (ECC P-256 or HSM capable of Ed25519)

## 1. Install Controllers
```
# AWS Load Balancer Controller
helm repo add eks https://aws.github.io/eks-charts
helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=<EKS_CLUSTER_NAME> \
  --set serviceAccount.create=false \
  --set region=<AWS_REGION>

# cert-manager
helm repo add jetstack https://charts.jetstack.io
helm upgrade --install cert-manager jetstack/cert-manager -n cert-manager --create-namespace \
  --set installCRDs=true

# external-dns (optional)
helm repo add bitnami https://charts.bitnami.com/bitnami
helm upgrade --install external-dns bitnami/external-dns -n external-dns --create-namespace \
  --set provider=aws --set policy=sync --set registry=txt \
  --set txtOwnerId=rtgf --set domainFilters[0]=example.com
```

## 2. TLS Certificates
Using cert-manager with Route53 DNS01:
```
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata: { name: letsencrypt-dns }
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef: { name: acme-key }
    solvers:
    - dns01:
        route53:
          region: <AWS_REGION>
          hostedZoneID: <ZONE_ID>
```
Then request certificate:
```
apiVersion: cert-manager.io/v1
kind: Certificate
metadata: { name: rtgf-tls, namespace: default }
spec:
  secretName: rtgf-tls
  dnsNames: [ "reg.example.com" ]
  issuerRef: { name: letsencrypt-dns, kind: ClusterIssuer }
```

## 3. Key Management
- **aws-kms (P-256)**: create an asymmetric KMS key (ECC_NIST_P256), grant IAM permissions, set `crypto.provider=aws-kms`, `crypto.awsKms.keyArn=<ARN>`, `crypto.awsKms.region=<AWS_REGION>`.
- **softhsm (Ed25519)**: create Kubernetes secret with sealed Ed25519 key material, reference via `crypto.softHsm.seedSecretRef`, rotate later to PQ/hybrid solution.

## 4. Ingress Annotations (values.yaml)
```
ingress:
  annotations:
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/ssl-redirect: "443"
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTPS":443}]'
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/backend-protocol: HTTP
```
Add WAF/Shield/access-log annotations as needed.

## 5. Install Chart
Prepare values (see `rtgf-registry/helm/values.yaml`) then deploy:
```
helm upgrade --install rtgf-registry rtgf-registry/helm \
  -f my-values.yaml \
  --set registry.domain=reg.example.com \
  --set crypto.provider=aws-kms \
  --set crypto.awsKms.keyArn=arn:aws:kms:eu-west-1:123456789:key/abcd \
  --set crypto.awsKms.region=eu-west-1 \
  --set persistence.transparency.storageClass=gp3 \
  --set persistence.revocations.storageClass=gp3
```

## 6. Smoke Test
```
curl -s https://reg.example.com/.well-known/rtgf | jq .
curl -s https://reg.example.com/rmt/EU/ai -H 'Accept: application/imt-rmt+json' | jq .
curl -s https://reg.example.com/revocations | jq .
```

Ensure responses include signed payloads, freshness metadata, and expected HTTPS headers.
