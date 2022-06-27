# Tutorial

This tutorial will guide you to have a quick first try of the Harbor operator and the Harbor cluster deployed and managed by the operator.

## Learn how does it work

If you want to learn the overall design of the Harbor operator before starting the quick try, check the [design doc](./arch/arch.md).

## Install Harbor operator stack

Check the prerequisites and install the Harbor operator base on your needs by following the [installation guide](./installation/installation.md).

Before moving on, make sure the harbor operator is successfully deployed in the Kubernetes cluster.

```shell
~/harbor-operator$ kubectl get all -n harbor-operator-ns
NAME                                     READY   STATUS    RESTARTS   AGE
pod/harbor-operator-85c94454bb-vm4mk     1/1     Running   0          6m12s
pod/minio-operator-67ddbdd4d8-v6qkl      1/1     Running   0          6m12s
pod/minio-operator-67ddbdd4d8-vqvrp      1/1     Running   0          6m12s
pod/postgres-operator-7dfbcfdf6b-wdgj4   1/1     Running   0          6m12s
pod/redisoperator-6f758c79fc-gk2cm       1/1     Running   0          6m12s

NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/operator            ClusterIP   10.96.141.225   <none>        4222/TCP   6m12s
service/postgres-operator   ClusterIP   10.96.247.95    <none>        8080/TCP   6m12s
service/webhook-service     ClusterIP   10.96.49.170    <none>        443/TCP    6m12s

NAME                                READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/harbor-operator     1/1     1            1           6m12s
deployment.apps/minio-operator      2/2     2            2           6m12s
deployment.apps/postgres-operator   1/1     1            1           6m12s
deployment.apps/redisoperator       1/1     1            1           6m12s

NAME                                           DESIRED   CURRENT   READY   AGE
replicaset.apps/harbor-operator-85c94454bb     1         1         1       6m12s
replicaset.apps/minio-operator-67ddbdd4d8      2         2         2       6m12s
replicaset.apps/postgres-operator-7dfbcfdf6b   1         1         1       6m12s
replicaset.apps/redisoperator-6f758c79fc       1         1         1       6m12s
```

## Deploy Harbor cluster

To deploy a Harbor cluster, you need to prepare a deployment manifest first. For creating the deployment manifest, you can clone one from the sample manifests listed in the [manifests/samples](../manifests/samples) folder and do any modifications based on your needs, or create from scratch by following the [CRD spec](./CRD/custom-resource-definition.md).

Learn more about the sample manifests, you can check [manifests reference](./manifests-reference.md#manifestssamples).

**NOTES:**
>To allow the deployed Harbor cluster to be accessible outside the Kubernetes cluster, make sure the ingress hosts and host in the `externalURL` should be mapping with accessible IPs in the /etc/hosts (for local development environments) or can be resolved and accessible by DNS resolver.
>TIPS: for local development, some plan-domain services like `sub-domain.<IP>.nip.io` can be used to provide simple public accessible hosts.

**ATTENTIONS:**
>To deploy Harbor cluster, you have to make sure a [default ingress class](https://kubernetes.io/docs/concepts/services-networking/ingress/#default-ingress-class) is marked in your cluster, otherwise, you need to explicitly set the [ingress class](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class) through the `ingressClassName` field of the `ingress` configuration. e.g. `spec.expose.core.ingress.ingressClassName`. For details, you can refer to the [CRD spec](./CRD/custom-resource-definition.md).

Here we clone the [full stack sample manifest](../manifests/samples/full_stack.yaml) as an example and modify the external host and ingress hosts with `sub-domain.<IP>.nip.io` pattern. Modified content is shown as below. Please pay attention here, the 'namespace', 'admin core secret', 'minio access secret' and 'cert-manager issuer/certificate/key' are pre-defined resources and bound to the deploying Harbor cluster.

`my_full_stack.yaml`:

```yaml
# Sample namespace
apiVersion: v1
kind: Namespace
metadata:
  name: cluster-sample-ns
---
# A secret of harbor admin password.
apiVersion: v1
kind: Secret
metadata:
  name: admin-core-secret
  namespace: cluster-sample-ns
data:
  secret: SGFyYm9yMTIzNDU=
type: Opaque
---
# A secret for minIO access.
apiVersion: v1
kind: Secret
metadata:
  name: minio-access-secret
  namespace: cluster-sample-ns
data:
  accesskey: YWRtaW4=
  secretkey: bWluaW8xMjM=
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: harbor-test-ca
  namespace: cluster-sample-ns
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZpekNDQTNPZ0F3SUJBZ0lVU2Fva1FldWczaGJHQVF3U1BqaklsMmV1akl3d0RRWUpLb1pJaHZjTkFRRUwKQlFBd1ZERUxNQWtHQTFVRUJoTUNRMDR4RERBS0JnTlZCQWdNQTFCRlN6RVJNQThHQTFVRUJ3d0lRbVZwSUVwcApibWN4RVRBUEJnTlZCQW9NQ0dkdmFHRnlZbTl5TVJFd0R3WURWUVFEREFoSVlYSmliM0pEUVRBZ0Z3MHlNREE0Ck1USXhNRFV5TkROYUdBOHlNVEl3TURjeE9URXdOVEkwTTFvd1ZERUxNQWtHQTFVRUJoTUNRMDR4RERBS0JnTlYKQkFnTUExQkZTekVSTUE4R0ExVUVCd3dJUW1WcElFcHBibWN4RVRBUEJnTlZCQW9NQ0dkdmFHRnlZbTl5TVJFdwpEd1lEVlFRRERBaElZWEppYjNKRFFUQ0NBaUl3RFFZSktvWklodmNOQVFFQkJRQURnZ0lQQURDQ0Fnb0NnZ0lCCkFPWlJaWWtGbVZCTmFiRVU0Y0RhcWEyN2s1K091VUhnMW9vU2NDbmJiQkZ4TkIyMi83Q3pYSVdPV21GQ1ZQalgKTHdBVjEzdTZwUWtNTWFqRHFQQS83bGR6OGFqWW05VlREZzgvcUdib2E2YW9OTzRHTExkeWlsaTY2R0xtSHBXawpCNi9KWjNKYUV2cHozOTQ4dk1pb0dDdXBjQnFFeGw4Z0hxcmdzeWpSOXVlcVM5bGU2Ni9nMWdEcVZtU0FVM3BQCndLa3dkUk1PaVFzeFNMOVRZL0lYYkNkVWVwc2xJQ2VHTEdqdjQvVUN0ZDlZVWhpSVJWOTdLc2dLRG4wQk8yOWMKVitXM0lDcW1yUDVsUUNYeUJQL0lPZW4zc0VoSlBDVTFuZUNZOWM2ODdKTlJJQ2VEWGUyTGJSUnRBcUIvblNXLwp1OGNzY3kwWFd6QjhicHN2dmU2d0hOMmNCU2thTXQ5MGN2N1JxMlpLNnRqMTltcHM1NXhoSjJibDFMTkNKbDZiClU2U0YvNXphZjQvdkF6SVJoUlpyMkJhQzlvdGc5U2ZQMXl2VHZBbVpUL1Z5T2h4R05QdGpZdU5rZm9YNVVSQ2QKWTB3N2l0S0VEZk82MDVJaXN1TEpGdUlhWVNMTzg1YWJSNmo3QzdNQnlWWHpxVEU1aGF1dFZlWno2SHNNOWxrbwp1dlQxejRZWm9hTVQrME4yT0dOcGtRdnlMTmZLM2djVitya0hkNURBSmdKWG1nNlVpZXh3UUtJZFIxNi9keWRPCkQ3RHpKTDVTU0tFbnNCaDA1NnRqQzU4NEdhVmljQ1U1WVBVdVBpSGx3dkRSREVLQytFNDQwdUxvQjUwTXE0TDgKOFI5Y3JZYklsWFkvdFF5eVdGL1FEbXFyZzMxeTVJRjBwV3JoOThoWm92eUxBZ01CQUFHalV6QlJNQjBHQTFVZApEZ1FXQkJTMUo2V3dyV1JPZlZ6UkNnVWtQeGVmMzV1ZW56QWZCZ05WSFNNRUdEQVdnQlMxSjZXd3JXUk9mVnpSCkNnVWtQeGVmMzV1ZW56QVBCZ05WSFJNQkFmOEVCVEFEQVFIL01BMEdDU3FHU0liM0RRRUJDd1VBQTRJQ0FRQ2IKR2RvUytOQmQ5KzJjUDViS1FFYjdsZXhXdEI5M1oyVWVsVkhhNllSRW90K2d1dU1yVkJnTGJvWmNGMS9pZzJIRApNV2l1bFIvQ2ZDN2tQbUxTWVZLRFlXWmJXRXBaUWhTNnYrSnlaQUkxY1lsU0d0cWVMcW41ZHBpUXZ2OTl2anJyCmFYTzV3a2U0Q3VZNnRqVTZzTkNhc2k5SDVXN09JRTlCRmxhSmtJeTh4SjNBeGpESlpDakthK2FsUFd0SnhGSSsKTHNKZkRmTFJTeERONlhPQ1hJTTZQM0NNVnYzemNudU9iQ3pwNUNKZWlULzA0eXdqWkJKQjB3bERVeGs3byt5WApjTzJUUEc4N01LOG1WelE3L1prZzgvUVZzTlNHcGlsSTBLYzVmNEtLR0NwUkR1RUx1dVNRbHNIdnZNaWRRcmQxCktDZHRHeDVIRFlJeDgrd0Z1RU5XTHZDVG4vd2dQQWtGbitqaU9Yb29CYmM4WllSU1NheW5mZzFtd0tpR2hPQ3EKK0hWV3hzdVNQSzFSZ2drTTFhMDJqTDZXby9wb0lsdkxvRFNhNjZDNHJoRnZmZGZWSlRJZWY2ZVBGb3k1Y0pNcwo4ZkZETU9SZ0Z4T29xbHliWGRyRmJsUEE0VzZUMjhLR0VwSmcyWVlDbmR3YkY3M0N6dkUrb2tlM254Uzcvakt4ClR4SFZJU01uY1ZONjNybnJoS1hNZEtCUzZvR1ErRkhXY0ZUNFhrbTBoU2drVURVTnZIMkxzaFUxREl5NGFrTmkKKzRhVlExdHlSd2huRjcrNVlUTDZnRTcremtZTXNOdkU2bDlFOU1MYW5TeUU5bnhPSDRKOHc0VUFuVTF2ZHZpQwpwQitXNXhsdVpvczZjMnUwbVk0d0UvdmVXV0lXRFdXdDFlT3J1RlBYUGc9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUpRd0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQ1Mwd2dna3BBZ0VBQW9JQ0FRRG1VV1dKQlpsUVRXbXgKRk9IQTJxbXR1NU9manJsQjROYUtFbkFwMjJ3UmNUUWR0dit3czF5RmpscGhRbFQ0MXk4QUZkZDd1cVVKRERHbwp3Nmp3UCs1WGMvR28ySnZWVXc0UFA2aG02R3VtcURUdUJpeTNjb3BZdXVoaTVoNlZwQWV2eVdkeVdoTDZjOS9lClBMeklxQmdycVhBYWhNWmZJQjZxNExNbzBmYm5xa3ZaWHV1djROWUE2bFprZ0ZONlQ4Q3BNSFVURG9rTE1VaS8KVTJQeUYyd25WSHFiSlNBbmhpeG83K1AxQXJYZldGSVlpRVZmZXlySUNnNTlBVHR2WEZmbHR5QXFwcXorWlVBbAo4Z1QveURucDk3QklTVHdsTlozZ21QWE92T3lUVVNBbmcxM3RpMjBVYlFLZ2Y1MGx2N3ZITEhNdEYxc3dmRzZiCkw3M3VzQnpkbkFVcEdqTGZkSEwrMGF0bVN1clk5ZlpxYk9lY1lTZG01ZFN6UWlaZW0xT2toZitjMm4rUDd3TXkKRVlVV2E5Z1dndmFMWVBVbno5Y3IwN3dKbVUvMWNqb2NSalQ3WTJMalpINkYrVkVRbldOTU80clNoQTN6dXRPUwpJckxpeVJiaUdtRWl6dk9XbTBlbyt3dXpBY2xWODZreE9ZV3JyVlhtYytoN0RQWlpLTHIwOWMrR0dhR2pFL3RECmRqaGphWkVMOGl6WHl0NEhGZnE1QjNlUXdDWUNWNW9PbEluc2NFQ2lIVWRldjNjblRnK3c4eVMrVWtpaEo3QVkKZE9lcll3dWZPQm1sWW5BbE9XRDFMajRoNWNMdzBReENndmhPT05MaTZBZWRES3VDL1BFZlhLMkd5SlYyUDdVTQpzbGhmMEE1cXE0TjljdVNCZEtWcTRmZklXYUw4aXdJREFRQUJBb0lDQUV6YnNOUnU1K0NpVkxqaFRReThhNDhzClgzRUpnY3o0S04vZWswdUVpNld1YjBQVFE3UkZ4b1JUSXRuOTlya3JwZVdUWkZ0SHg3Y2pPSmNtNUFONGNpTUEKOEEzMmF0cGZZdnUzdEl6UzFzbkFyQmthT21YbGRVRnk3Z1hDNFVYeWZSWXVVYlVaVmVmNkx5VE1nL3M2RFFiVQovakg3U08rSm1uSlBsYm56aHo5NzF0L3RDeDJnSEFvbUtUcFVrSWJxZ2xKemR6NHF4WlRVbDRBeFpkTHQrZ3VOCjUzUktpVlpuTWY2NnZ3bU9JLzhxVEFzZnZuYkVkVnhYN3NuTVZYY3VDNjcrMDE4b1MrYUJCMDBpWElTMjNveXoKT1VLR0hlb1U0R0NJNnM1WXdXSFAycmtVMzQxYno4VFhNOTgzZHN1WUZpTzdNNXhDaFEzREdHMzFHcDdDYW43dgpaeWVCaGwxeXBvZHVVYjQxUW0vRzB4S3gyelRyeWR6RXBOZUNidCtJazhjMDBvMW1JTXRVdkc3ekVzeFFBc2M2CnBGMTZEK0lEZ01UdzZLM1FuYmpiMzE2QlFxQzhoOW9PMkJYZ1c0UGdnVVphWjByS2RTMGs2Ylo1dnE3T3I2Qk4KZW8zamlwbjlxZWxlODZHaU5aMmF5cmNyck50M2VlSkdmK3h0ZXBMVERYU2krTCt2OS8rd056bFZ1Vm5VWVIyWQpLdC84Y1htR2tvcmtDK2djN0RYbnhiMllaRTE2RDZWcE5rcEd0QlZpQTVZMXdDeGdtcVNRbDRneGNMZWE2RUNWCkVCeFhqOWFvWjZJTGlYRnRwRFZtVStVVVZZNmVPMSsxSVBqNmFiNEdibEhPRE5BcFJGekpUS29oS2tBY0ViZXkKNi90NW1KNGRjZXAxRm5veTFDQXhBb0lCQVFENXk4NU4xQmgwMCtwdmpEU3VWdkwrcll4VlFZakdTLzFtWlYvYgorSXo3N3hBOTlzU2ZkSzNWS2UxSk15bFg0Zkx4SEo2TG9kYUxoMysyU3U3MTNqSTQzUTJ4bFRiSURJeFQxdEJECjhucEpzNHI2YktINHh2NGYyby9RYzhnWjQwTEg4bkdTc1VPZUdEYU1DTlN3ZjNxR0FnMFlJY2EzSkhQUnJBdk8KZE9OZmFicWVyaVFYcUt4VzFObUxhdGtaVTI0S0twWmcra0oyV2picGl3dms0SFM4Zmdwb0puRkd2L3I5ZEJnQgprQW8yeGpTOEJMZFU5cno2SjFpQ1I2ZURPN3dhZCtiWmlIVFVBSXFyUmRKYU5lSmFTS2dwYUpheEp6cUNQODBkCjd4LzFKQTBSY3NvUU5VMHVpeml5MmdKUCt4bVgvN2FMNzJCdDBJYzZtT1VicTlwSEFvSUJBUURzQ2IvVklLSUwKUWNrYlZpSW1DanFuQllVQW5RZmpRaUlpaTFqV3lCOGVvOXlqVVM2b0JNMTdOOVJrU1psVHZtMzBGcjdsVXAwKwpNUmVmVkxLdG5EYUhQOXZqWTVvQ0xObEs3Ryt1U0hZRTl3NFI0T2lzelYzeWl4L3JVV0cramVKRHhGeGVnUFJKCnc1cUExclZUblcvVGlMc0ltUG94RG1HRUNKRjBhTi83dmMvaW51ZS9LVGZSc2J6OThmTWd3SUltTHVBT0UrekwKMzRPRHdMRU1BbjZCaG5sTUlweE8xSDU1QUdVOEc0MnNjM1ZwRkcvMHBWYU14TzR5VnQzaDY3U2JJV05zT1hBYgpPdm43L04xVlQ5TUtWUkFVMTZ2ZnJENW82TjRPZnd0eW1TcVhHUFBhTXRBRDYwM0l6UWp2ajV2WCsvRDZEMVJSCjBSOEpNYzVxRVdtZEFvSUJBUURqRVdiSnpNRW1nZlNiemNHZHNTQldiZ0FoQjkrREVsU1luaEpUYlU4TFBMZHcKL0Q2a0RIWndUUnFMN2R2cExWV2Y0N29qaDh2MUxnamo5cDNlRmt0azhWeWZUdHByWXl5MGtaTGtFU2tra2ZjRgp5WFk3SlBpZ2tCY25EL2lYdjhSVzZZWmdLSThreVRIY2ZiS0pkbmcwRk8wK1FJWFl1V1FtOXRRTXFxaDlkU2pWClVjc3hUbnpLdWRXL0xET0pHQlB4WGVFdzZvMDc5S255QmhtYnhvV1hTcU8vSlNMWGczQnVzUGVaaEF3azJtdloKZGhnSlBmbHZGQkVhN0hQVGtadGVIQnhYSmZtOU5YallWREh4R3daVnQ3SlZZZU9KeWZVZnJVdVJxR3RPZGFVRApkV3RFN0k3cWZsZmVETnNKUldKd2oxeXJPOEJXVXJaNmg3M01ONTNGQW9JQkFRQ042OHJCTGc3Z3I5eG9xR0I5CitOYU5TRjlSSUJuM0JmT2FTNmpOODZQcWUySVZmS0dOK3QxR0FpcWRaamRmeC9jNnRWWndjajBEZ09jUU1SQUMKSFJRWVBFaE5MNzBSSThBL01XeHhJVFo5QThNYzh0dFQwMk55aXo1VThpalFOMlZkazdwcVJDVWVHUk5UOWtVdQprbElEb1ltN3dLZG1TWnhPbEF4Skx5bkZwcnBSSzNSeVZ5a3QxeTJvandxOW5hSmpyUG1nM1ZBYXdUakZSbDN0CnQ2NHkyUlVqdHdlK3lqdUZLN3l5NkdwRnoySkFIVDYzblpZdHE5Y0F3NFJENjhJN0tGY3NZbGpLdHFwS1hoOEMKeGExQjRDVjhNclV3RnRPcnBxQ2xuTFBZWXNuZDhlM2xPM29oY1NEaTVJMUQ1Vmd5QkZVL05XcGdpMW1hNEt5WQowQUZ0QW9JQkFGZ1h0ZWM2VE1EVGVpRDFXQzRFK3E5aHNKZHBSamFwTGJqOGlTdFVyZVNmL2xIcFQrb1dYdUt3CnJnS2d6bm8vVjNxeEdKaFBWaFJJUW1Kb3U2M01LY1hxWXVwSlkxRTdjMmVhbU84SkpPZlpNTFFrWlN1MWlxSmIKZ3JITCt6YVAvTjRqVmkrWTJqbG81OEloYzBLNzVtak9rYjVBZ2ExOFhySnBlYlRHbFdnWFlxY05xUmZzbWhVago3VmMxQk9vVGwwb3pPU21JRmpyQUp3R0pNZjFuQzBuODI5Mks2WERyN0tDbDJXazN2OC9oVzNpRzJ6cW55aFBHCkl3czY0L0NLYU03OTdIc0R1ZXlBTU5vUVJXUHk5d2I3YXZqZlI2TDNGOWczS2l6c1pVbjFtUWcwc0JsQVRTU2QKVEN6SGxtdjZqQVFqY0RxVzdxQ3I0SHRFTXJBNjJGWT0KLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo=
---
# Cert issuer
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: cluster-sample-ns
spec:
  ca:
    secretName: harbor-test-ca
---
# Certificates of ingress
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: sample-public-certificate
  namespace: cluster-sample-ns
spec:
  secretName: sample-public-certificate
  dnsNames:
    - core.10.10.10.100.nip.io
    - notary.10.10.10.100.nip.io
    - minio.10.10.10.100.nip.io
  issuerRef:
    name: selfsigned-issuer
    kind: Issuer
---
# Full stack Harbor
apiVersion: goharbor.io/v1beta1
kind: HarborCluster
metadata:
  name: harborcluster-sample
  namespace: cluster-sample-ns
spec:
  version: 2.3.0
  logLevel: info
  network:
    ipFamilies:
      - IPv4
      - IPv6
  imageSource:
    repository: ghcr.io/goharbor
  harborAdminPasswordRef: admin-core-secret
  externalURL: https://core.10.10.10.100.nip.io
  expose:
    core:
      ingress:
        host: core.10.10.10.100.nip.io
        controller: default
      tls:
        certificateRef: sample-public-certificate
    notary:
      ingress:
        host: notary.10.10.10.100.nip.io
        controller: default
      tls:
        certificateRef: sample-public-certificate
  internalTLS:
    enabled: true
  portal: {}
  registry:
    metrics:
      enabled: true
  core:
    tokenIssuer:
      name: selfsigned-issuer
      kind: Issuer
    metrics:
      enabled: true
  chartmuseum: {}
  exporter: {}
  trivy:
    skipUpdate: false
    storage: {}
  notary:
    migrationEnabled: true
  database:
    kind: Zlando/PostgreSQL
    spec:
      zlandoPostgreSql:
        operatorVersion: "1.5.0"
        storage: 1Gi
        replicas: 1
        resources:
          limits:
            cpu: 500m
            memory: 500Mi
          requests:
            cpu: 100m
            memory: 250Mi
  storage:
    kind: MinIO
    spec:
      minIO:
        operatorVersion: "4.0.6"
        replicas: 2
        secretRef: minio-access-secret
        redirect:
          enable: true
          expose:
            ingress:
              host: minio.10.10.10.100.nip.io
            tls:
              certificateRef: sample-public-certificate
        volumesPerServer: 2
        volumeClaimTemplate:
          spec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 10Gi
  cache:
    kind: RedisFailover
    spec:
      redisFailover:
        operatorVersion: "1.0.0"
        server:
          replicas: 1
          resources:
            limits:
              cpu: 500m
              memory: 500Mi
            requests:
              cpu: 100m
              memory: 250Mi
        sentinel:
          replicas: 1

```

>Notes: for using the `inClusterStorage` with `redirect` is `enable(=true)`, make sure the ingress route set in the `redirect.expose.ingress.host` field can be accessed from both inside and outside the cluster.

Apply the above modified deployment manifest to your cluster.

```shell
kubectl create -f my_full_stack.yaml
```

Check the overall status of the deployed Harbor cluster.

```shell
kubectl get harborcluster/harborcluster-sample -n cluster-sample-ns -o wide
```

The `name`, `public URL`, `status`, `operator version` and `operator commit` info are printed out:

```log
NAME                   PUBLIC URL                          STATUS    OPERATOR VERSION   OPERATOR GIT COMMIT
harborcluster-sample   https://core.10.10.10.100.nip.io   healthy   1.2.0          35056b427665e9ee6331e7b3ddede8c26b86dbc5
```

You can check more detailed status(conditions) info of the deployed Harbor cluster with appending `-o yaml`.

```shell
kubectl get harborcluster/harborcluster-sample -n cluster-sample-ns -o yaml
```

Some status info like the following data is printed out:

```yaml
status:
    conditions:
    - lastTransitionTime: "2022-01-18T10:02:02Z"
      status: "True"
      type: StorageReady
    - lastTransitionTime: "2022-01-18T10:00:55Z"
      message: Harbor component database secrets are already create
      reason: Database is ready
      status: "True"
      type: DatabaseReady
    - lastTransitionTime: "2022-01-18T10:00:27Z"
      message: harbor component redis secrets are already create.
      reason: redis already ready
      status: "True"
      type: CacheReady
    - status: "False"
      type: InProgress
    - lastTransitionTime: "2022-01-18T10:02:31Z"
      status: "True"
      type: ServiceReady
    - status: "False"
      type: Failed
    observedGeneration: 1
    operator:
      controllerGitCommit: 35056b427665e9ee6331e7b3ddede8c26b86dbc5
      controllerName: harborcluster
      controllerVersion: dev_master
    revision: 1642500151584677122
    status: healthy
```

You can also check what Kubernetes resources are created by getting all.

```shell
kubectl get all -n cluster-sample-ns
```

A few of resources info like the following data are output:

```log
NAME                                                                  READY   STATUS      RESTARTS   AGE
pod/harborcluster-sample-harbor-harbor-chartmuseum-5b68bd46b8-lq9g9   1/1     Running     0          105s
pod/harborcluster-sample-harbor-harbor-core-7849f9c844-vf4w7          1/1     Running     0          105s
pod/harborcluster-sample-harbor-harbor-exporter-dd5c99bc9-br9sk       1/1     Running     0          98s
pod/harborcluster-sample-harbor-harbor-jobservice-668dd85cc4-2q9bf    1/1     Running     0          98s
pod/harborcluster-sample-harbor-harbor-notaryserver-77779bf8bf8t5tv   1/1     Running     0          108s
pod/harborcluster-sample-harbor-harbor-notarysigner-5c88dc99d8d5s64   1/1     Running     0          113s
pod/harborcluster-sample-harbor-harbor-portal-64465c4954-52zxc        1/1     Running     0          109s
pod/harborcluster-sample-harbor-harbor-registry-756dbcf5bb-zcrfn      2/2     Running     0          104s
pod/harborcluster-sample-harbor-harbor-trivy-599ff5c789-hcblv         1/1     Running     0          108s
pod/minio-harborcluster-sample--1-4n72x                               0/1     Completed   0          117s
pod/minio-harborcluster-sample-zone-harbor-0                          1/1     Running     0          3m23s
pod/minio-harborcluster-sample-zone-harbor-1                          1/1     Running     0          3m23s
pod/postgresql-cluster-sample-ns-harborcluster-sample-0               1/1     Running     0          3m31s
pod/rfr-harborcluster-sample-redis-0                                  1/1     Running     0          3m31s
pod/rfs-harborcluster-sample-redis-6fddf664-9nwsc                     1/1     Running     0          3m31s

NAME                                                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)            AGE
service/harborcluster-sample-harbor-harbor-chartmuseum             ClusterIP   10.96.9.53      <none>        443/TCP            105s
service/harborcluster-sample-harbor-harbor-core                    ClusterIP   10.96.175.82    <none>        443/TCP,8001/TCP   106s
service/harborcluster-sample-harbor-harbor-exporter                ClusterIP   10.96.1.77      <none>        8001/TCP           98s
service/harborcluster-sample-harbor-harbor-jobservice              ClusterIP   10.96.67.178    <none>        443/TCP            98s
service/harborcluster-sample-harbor-harbor-notaryserver            ClusterIP   10.96.235.154   <none>        443/TCP            108s
service/harborcluster-sample-harbor-harbor-notarysigner            ClusterIP   10.96.246.138   <none>        7899/TCP           113s
service/harborcluster-sample-harbor-harbor-portal                  ClusterIP   10.96.177.105   <none>        443/TCP            109s
service/harborcluster-sample-harbor-harbor-registry                ClusterIP   10.96.24.80     <none>        443/TCP,8001/TCP   105s
service/harborcluster-sample-harbor-harbor-registryctl             ClusterIP   10.96.171.157   <none>        443/TCP            105s
service/harborcluster-sample-harbor-harbor-trivy                   ClusterIP   10.96.163.122   <none>        443/TCP            108s
service/minio                                                      ClusterIP   10.96.89.191    <none>        80/TCP             3m25s
service/minio-harborcluster-sample-console                         ClusterIP   10.96.29.2      <none>        9090/TCP           3m24s
service/minio-harborcluster-sample-hl                              ClusterIP   None            <none>        9000/TCP           3m23s
service/postgresql-cluster-sample-ns-harborcluster-sample          ClusterIP   10.96.16.82     <none>        5432/TCP           3m31s
service/postgresql-cluster-sample-ns-harborcluster-sample-config   ClusterIP   None            <none>        <none>             3m23s
service/postgresql-cluster-sample-ns-harborcluster-sample-repl     ClusterIP   10.96.33.5      <none>        5432/TCP           3m31s
service/rfs-harborcluster-sample-redis                             ClusterIP   10.96.53.11     <none>        26379/TCP          3m31s

NAME                                                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/harborcluster-sample-harbor-harbor-chartmuseum    1/1     1            1           105s
deployment.apps/harborcluster-sample-harbor-harbor-core           1/1     1            1           105s
deployment.apps/harborcluster-sample-harbor-harbor-exporter       1/1     1            1           98s
deployment.apps/harborcluster-sample-harbor-harbor-jobservice     1/1     1            1           98s
deployment.apps/harborcluster-sample-harbor-harbor-notaryserver   1/1     1            1           108s
deployment.apps/harborcluster-sample-harbor-harbor-notarysigner   1/1     1            1           113s
deployment.apps/harborcluster-sample-harbor-harbor-portal         1/1     1            1           109s
deployment.apps/harborcluster-sample-harbor-harbor-registry       1/1     1            1           104s
deployment.apps/harborcluster-sample-harbor-harbor-trivy          1/1     1            1           108s
deployment.apps/rfs-harborcluster-sample-redis                    1/1     1            1           3m31s

NAME                                                                         DESIRED   CURRENT   READY   AGE
replicaset.apps/harborcluster-sample-harbor-harbor-chartmuseum-5b68bd46b8    1         1         1       105s
replicaset.apps/harborcluster-sample-harbor-harbor-core-7849f9c844           1         1         1       105s
replicaset.apps/harborcluster-sample-harbor-harbor-exporter-dd5c99bc9        1         1         1       98s
replicaset.apps/harborcluster-sample-harbor-harbor-jobservice-668dd85cc4     1         1         1       98s
replicaset.apps/harborcluster-sample-harbor-harbor-notaryserver-77779bf8bf   1         1         1       108s
replicaset.apps/harborcluster-sample-harbor-harbor-notarysigner-5c88dc99d8   1         1         1       113s
replicaset.apps/harborcluster-sample-harbor-harbor-portal-64465c4954         1         1         1       109s
replicaset.apps/harborcluster-sample-harbor-harbor-registry-756dbcf5bb       1         1         1       104s
replicaset.apps/harborcluster-sample-harbor-harbor-trivy-599ff5c789          1         1         1       108s
replicaset.apps/rfs-harborcluster-sample-redis-6fddf664                      1         1         1       3m31s

NAME                                                                 READY   AGE
statefulset.apps/minio-harborcluster-sample-zone-harbor              2/2     3m23s
statefulset.apps/postgresql-cluster-sample-ns-harborcluster-sample   1/1     3m31s
statefulset.apps/rfr-harborcluster-sample-redis                      1/1     3m31s

NAME                                   COMPLETIONS   DURATION   AGE
job.batch/minio-harborcluster-sample   1/1           1s         117s

NAME                                                                         TEAM                           VERSION   PODS   VOLUME   CPU-REQUEST   MEMORY-REQUEST   AGE     STATUS
postgresql.acid.zalan.do/postgresql-cluster-sample-ns-harborcluster-sample   postgresql-cluster-sample-ns   12        1      1Gi      100m          250Mi            3m31s   Running

NAME                                                               AGE
redisfailover.databases.spotahome.com/harborcluster-sample-redis   3m31s
```

Of course, you can also check other resources such as `ingress`, `secret`, `pv`, `certificate` and `configMap` etc. under the specified namespace with `kubectl get xxxx -n cluster-sample-ns` commands.

## Try the deployed Harbor cluster [Optional]

Now you can try the deployed Harbor cluster.

1. Navigate to the Harbor home address `https://core.10.10.10.100.nip.io` and login Harbor with the root user `admin` and the password you provided in the deployment manifest above.

    >In case you forgot the password, try to get it with the command:

    ```bash
    kubectl get secret \
    "$(kubectl get harborcluster harborcluster-sample -n cluster-sample-ns -o jsonpath='{.spec.harborAdminPasswordRef}')" \
    -n cluster-sample-ns \
    -o jsonpath='{.data.secret}' \
    | base64 --decode
    ```

1. Navigate to the `Projects` list page and click `NEW PROJECT` button to create a new project named `my-harbor`.
1. Open the `Registries` management view that is under `Administration` part and click the `NEW ENDPOINT` button to open the `New Registry Endpoint` dialog.
1. Select `Docker Hub` as the `Provider` and input `myhub` as the name of this creating endpoint. Test the connection by clicking `TEST CONNECTION` at the bottom and then click `OK` button to create it after connection health is confirmed.
1. Open the `Replications` management view that is under `Administration` part and click `NEW REPLICATION RULE` button to open the `New Replication Rule` dialog.
1. Input `my-first-rule` as the name of the creating replication rule. Select `Pull-based` radio to switch to pull mode. For the `Source resource filter` section, input `goharbor/harbor-core` in the `Name` field and `**` in the `Tag` field.
1. Select `myhub-https://hub.docker.com` as the `Source registry`.
1. Input `my-harbor` in the `Destination namespace`.
1. Keep `Manual` in the `Trigger mode`.
1. Click `SAVE` to persist the replication rule.
1. Select the replication rule `my-first-rule` in the list and click the `REPLICATION` button to start execute the replication process defined by the selected rule.
1. Check the overall status from the created execution with ID `1` after the replication process is started.
1. Navigate back to the project `my-harbor` and you'll see the related artifacts have been already put under the project once the replication process is successfully completed.

For fully using the deployed Harbor cluster, you have to get the certificate authority used to generate the public certificate and install it on your computer (on the system scope, docker daemon + browser).

```shell
kubectl get secret sample-public-certificate -n cluster-sample-ns \
-o jsonpath='{.data.ca\.crt}' \
| base64 --decode
```

Of course, you can try other operations of Harbor like scanning etc.
