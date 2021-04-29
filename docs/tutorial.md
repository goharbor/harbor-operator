# Tutorial

This tutorial will guides you to have a quick first try of the Harbor operator and the Harbor cluster deployed and managed by the operator.

## Learn how does it work

If you want to learn the overall design of the Harbor operator before starting the quick try, check the [design doc](./arch/arch.md).

## Install Harbor operator stack

Check the prerequisites and install the Harbor operator base on your needs by following the [installation guide](./installation/installation.md).

Before moving on, make sure the harbor operator is successfully deployed in the Kubenetes cluster.

```shell
~/harbor-operator$ k8s get all -n harbor-operator-ns
NAME                                    READY   STATUS    RESTARTS   AGE
pod/harbor-operator-54454997d-bjkt9     1/1     Running   0          59s
pod/minio-operator-c4d8f7b4d-dztwl      1/1     Running   0          59s
pod/postgres-operator-94578ffd5-6kdql   1/1     Running   0          58s
pod/redisoperator-6b75fc4555-ps5kj      1/1     Running   0          58s

NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/operator            ClusterIP   10.96.114.34    <none>        4222/TCP,4233/TCP   59s
service/postgres-operator   ClusterIP   10.96.208.57    <none>        8080/TCP            59s
service/webhook-service     ClusterIP   10.96.234.182   <none>        443/TCP             59s

NAME                                READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/harbor-operator     1/1     1            1           59s
deployment.apps/minio-operator      1/1     1            1           59s
deployment.apps/postgres-operator   1/1     1            1           59s
deployment.apps/redisoperator       1/1     1            1           58s

NAME                                          DESIRED   CURRENT   READY   AGE
replicaset.apps/harbor-operator-54454997d     1         1         1       59s
replicaset.apps/minio-operator-c4d8f7b4d      1         1         1       59s
replicaset.apps/postgres-operator-94578ffd5   1         1         1       58s
replicaset.apps/redisoperator-6b75fc4555      1         1         1       58s
```

## Deploy Harbor cluster

To deploy a Harbor cluster, you need to prepare a deployment manifest first. For creating the deployment manifest, you can clone one from the sample manifests listed in the [manifests/samples](../manifests/samples) folder and do any modifications based on your needs, or create from scratch by following the [CRD spec](./CRD/custom-resource-definition.md).

Learn more about the sample manifests, you can check [manifests reference](./manifests-reference.md#manifestssamples).

> NOTES: to allow the deployed Harbor cluster to be accessible outside the Kubenetes cluster, make sure the ingress hosts and host in the `externalURL` should be mapping with accessible IPs in the /etc/hosts (for local development environments) or can be resolved and accessible by DNS resolver.
>TIPS: for local development, some plan-domain services like `sub-domain.<IP>.xip.io` can be used to provide simple public accessible hosts.

Here we clone the [full stack sample manifest](../manifests/samples/full_stack.yaml) as an example and modify the external host and ingress hosts with `sub-domain.<IP>.xip.io` pattern. Modified content is shown as below. Please pay attention here, the 'namespace', 'admin password', 'minio access secret' and 'cert-manager issuer' are pre-defined resources and bound to the deploying Harbor cluster.

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
    - core.10.10.10.100.xip.io
    - notary.10.10.10.100.xip.io
    - minio.10.10.10.100.xip.io
  issuerRef:
    name: selfsigned-issuer
    kind: Issuer
---
# Full stack Harbor
apiVersion: goharbor.io/v1alpha3
kind: HarborCluster
metadata:
  name: harborcluster-sample
  namespace: cluster-sample-ns
spec:
  logLevel: info
  imageSource:
    repository: ghcr.io/goharbor
  harborAdminPasswordRef: admin-core-secret
  externalURL: https://core.10.10.10.100.xip.io
  expose:
    core:
      ingress:
        host: core.10.10.10.100.xip.io
      tls:
        certificateRef: sample-public-certificate
    notary:
      ingress:
        host: notary.10.10.10.100.xip.io
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
  inClusterDatabase:
    kind: PostgresSQL
    postgresSqlSpec:
      storage: 1Gi
      replicas: 1
      resources:
        limits:
          cpu: 500m
          memory: 500Mi
        requests:
          cpu: 100m
          memory: 250Mi
  inClusterStorage:
    kind: MinIO
    minIOSpec:
      replicas: 2
      secretRef: minio-access-secret
      redirect:
        enable: true
        expose:
          ingress:
            host: minio.10.10.10.100.xip.io
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
  inClusterCache:
    kind: Redis
    redisSpec:
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
harborcluster-sample   https://core.10.10.10.100.xip.io   healthy   0.0.0-dev          796c6d3acd4a8801f77b242c2797f931944656c6
```

You can check more detailed status(conditions) info of the deployed Harbor cluster with appending `-o yaml`.

```shell
kubectl get harborcluster/harborcluster-sample -n cluster-sample-ns -o yaml
```

Some status info like the following data is printed out:

```yaml
status:
  conditions:
  - lastTransitionTime: "2021-04-20T07:13:02Z"
    status: "True"
    type: StorageReady
  - lastTransitionTime: "2021-04-20T07:15:38Z"
    message: Harbor component database secrets are already create
    reason: Database is ready
    status: "True"
    type: DatabaseReady
  - lastTransitionTime: "2021-04-20T07:15:38Z"
    message: harbor component redis secrets are already create.
    reason: redis already ready
    status: "True"
    type: CacheReady
  - status: "False"
    type: InProgress
  - lastTransitionTime: "2021-04-20T07:16:05Z"
    status: "True"
    type: ServiceReady
```

You can also check what Kubernetes resources are created by getting all.

```shell
kubectl get all -n cluster-sample-ns
```

A few of resources info like the following data are output:

```log
NAME                                                              READY   STATUS    RESTARTS   AGE
harborcluster-sample-harbor-harbor-chartmuseum-7b75745c8f-dvzh9   1/1     Running   0          9m56s
harborcluster-sample-harbor-harbor-core-8446dc9bdc-t8jms          1/1     Running   0          10m
harborcluster-sample-harbor-harbor-exporter-5bc9ccf9db-c9qsl      1/1     Running   0          8m24s
harborcluster-sample-harbor-harbor-jobservice-6586c85665-dsx99    1/1     Running   0          8m23s
harborcluster-sample-harbor-harbor-notaryserver-78c799554-pdbmd   1/1     Running   0          10m
harborcluster-sample-harbor-harbor-notarysigner-cc457d87d-556zm   1/1     Running   0          10m
harborcluster-sample-harbor-harbor-portal-6599969476-lmppn        1/1     Running   0          10m
harborcluster-sample-harbor-harbor-registry-7558db475-n78fm       1/1     Running   0          9m59s
harborcluster-sample-harbor-harbor-registryctl-7d5479c4f8-czk8f   1/1     Running   0          8m40s
harborcluster-sample-harbor-harbor-trivy-678966b955-pmb9b         1/1     Running   0          10m
minio-harborcluster-sample-zone-harbor-0                          1/1     Running   0          11m
minio-harborcluster-sample-zone-harbor-1                          1/1     Running   0          11m
postgresql-cluster-sample-ns-harborcluster-sample-0               1/1     Running   0          11m
rfr-harborcluster-sample-redis-0                                  1/1     Running   0          11m
rfs-harborcluster-sample-redis-6b7f4c4756-mrpsz                   1/1     Running   0          11m
steven@steven-zou:~/samples$ k8s get all -n cluster-sample-ns
NAME                                                                  READY   STATUS    RESTARTS   AGE
pod/harborcluster-sample-harbor-harbor-chartmuseum-7b75745c8f-dvzh9   1/1     Running   0          21m
pod/harborcluster-sample-harbor-harbor-core-8446dc9bdc-t8jms          1/1     Running   0          21m
pod/harborcluster-sample-harbor-harbor-exporter-5bc9ccf9db-c9qsl      1/1     Running   0          19m
pod/harborcluster-sample-harbor-harbor-jobservice-6586c85665-dsx99    1/1     Running   0          19m
pod/harborcluster-sample-harbor-harbor-notaryserver-78c799554-pdbmd   1/1     Running   0          21m
pod/harborcluster-sample-harbor-harbor-notarysigner-cc457d87d-556zm   1/1     Running   0          21m
pod/harborcluster-sample-harbor-harbor-portal-6599969476-lmppn        1/1     Running   0          21m
pod/harborcluster-sample-harbor-harbor-registry-7558db475-n78fm       1/1     Running   0          21m
pod/harborcluster-sample-harbor-harbor-registryctl-7d5479c4f8-czk8f   1/1     Running   0          20m
pod/harborcluster-sample-harbor-harbor-trivy-678966b955-pmb9b         1/1     Running   0          21m
pod/minio-harborcluster-sample-zone-harbor-0                          1/1     Running   0          22m
pod/minio-harborcluster-sample-zone-harbor-1                          1/1     Running   0          22m
pod/postgresql-cluster-sample-ns-harborcluster-sample-0               1/1     Running   0          23m
pod/rfr-harborcluster-sample-redis-0                                  1/1     Running   0          23m
pod/rfs-harborcluster-sample-redis-6b7f4c4756-mrpsz                   1/1     Running   0          23m

NAME                                                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)            AGE
service/harborcluster-sample-harbor-harbor-chartmuseum             ClusterIP   10.96.177.5     <none>        443/TCP            21m
service/harborcluster-sample-harbor-harbor-core                    ClusterIP   10.96.179.126   <none>        443/TCP,8001/TCP   21m
service/harborcluster-sample-harbor-harbor-exporter                ClusterIP   10.96.34.119    <none>        8001/TCP           19m
service/harborcluster-sample-harbor-harbor-jobservice              ClusterIP   10.96.55.96     <none>        443/TCP            19m
service/harborcluster-sample-harbor-harbor-notaryserver            ClusterIP   10.96.222.39    <none>        443/TCP            21m
service/harborcluster-sample-harbor-harbor-notarysigner            ClusterIP   10.96.108.105   <none>        7899/TCP           21m
service/harborcluster-sample-harbor-harbor-portal                  ClusterIP   10.96.236.44    <none>        443/TCP            21m
service/harborcluster-sample-harbor-harbor-registry                ClusterIP   10.96.44.59     <none>        443/TCP,8001/TCP   21m
service/harborcluster-sample-harbor-harbor-registryctl             ClusterIP   10.96.113.239   <none>        443/TCP            20m
service/harborcluster-sample-harbor-harbor-trivy                   ClusterIP   10.96.142.63    <none>        443/TCP            21m
service/minio                                                      ClusterIP   10.96.149.137   <none>        80/TCP             23m
service/minio-harborcluster-sample-hl                              ClusterIP   None            <none>        9000/TCP           23m
service/postgresql-cluster-sample-ns-harborcluster-sample          ClusterIP   10.96.138.163   <none>        5432/TCP           23m
service/postgresql-cluster-sample-ns-harborcluster-sample-config   ClusterIP   None            <none>        <none>             22m
service/postgresql-cluster-sample-ns-harborcluster-sample-repl     ClusterIP   10.96.197.105   <none>        5432/TCP           23m
service/rfs-harborcluster-sample-redis                             ClusterIP   10.96.250.26    <none>        26379/TCP          23m

NAME                                                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/harborcluster-sample-harbor-harbor-chartmuseum    1/1     1            1           21m
deployment.apps/harborcluster-sample-harbor-harbor-core           1/1     1            1           21m
deployment.apps/harborcluster-sample-harbor-harbor-exporter       1/1     1            1           19m
deployment.apps/harborcluster-sample-harbor-harbor-jobservice     1/1     1            1           19m
deployment.apps/harborcluster-sample-harbor-harbor-notaryserver   1/1     1            1           21m
deployment.apps/harborcluster-sample-harbor-harbor-notarysigner   1/1     1            1           21m
deployment.apps/harborcluster-sample-harbor-harbor-portal         1/1     1            1           21m
deployment.apps/harborcluster-sample-harbor-harbor-registry       1/1     1            1           21m
deployment.apps/harborcluster-sample-harbor-harbor-registryctl    1/1     1            1           20m
deployment.apps/harborcluster-sample-harbor-harbor-trivy          1/1     1            1           21m
deployment.apps/rfs-harborcluster-sample-redis                    1/1     1            1           23m

NAME                                                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/harborcluster-sample-harbor-harbor-chartmuseum-7b75745c8f   1         1         1       21m
replicaset.apps/harborcluster-sample-harbor-harbor-core-8446dc9bdc          1         1         1       21m
replicaset.apps/harborcluster-sample-harbor-harbor-exporter-5bc9ccf9db      1         1         1       19m
replicaset.apps/harborcluster-sample-harbor-harbor-jobservice-6586c85665    1         1         1       19m
replicaset.apps/harborcluster-sample-harbor-harbor-notaryserver-78c799554   1         1         1       21m
replicaset.apps/harborcluster-sample-harbor-harbor-notarysigner-cc457d87d   1         1         1       21m
replicaset.apps/harborcluster-sample-harbor-harbor-portal-6599969476        1         1         1       21m
replicaset.apps/harborcluster-sample-harbor-harbor-registry-7558db475       1         1         1       21m
replicaset.apps/harborcluster-sample-harbor-harbor-registryctl-7d5479c4f8   1         1         1       20m
replicaset.apps/harborcluster-sample-harbor-harbor-trivy-678966b955         1         1         1       21m
replicaset.apps/rfs-harborcluster-sample-redis-6b7f4c4756                   1         1         1       23m

NAME                                                                 READY   AGE
statefulset.apps/minio-harborcluster-sample-zone-harbor              2/2     22m
statefulset.apps/postgresql-cluster-sample-ns-harborcluster-sample   1/1     23m
statefulset.apps/rfr-harborcluster-sample-redis                      1/1     23m

NAME                                                                         TEAM                           VERSION   PODS   VOLUME   CPU-REQUEST   MEMORY-REQUEST   AGE   STATUS
postgresql.acid.zalan.do/postgresql-cluster-sample-ns-harborcluster-sample   postgresql-cluster-sample-ns   12        1      1Gi      100m          250Mi            23m   Running

NAME                                                               AGE
redisfailover.databases.spotahome.com/harborcluster-sample-redis   23m
```

Of course, you can also check other resources such as `secret`, `pv`, `certificate` and `configMap` etc. under the specified namespace with `kubectl get xxxx -n cluster-sample-ns` commands.

## Try the deployed Harbor cluster [Optional]

Now you can try the deployed Harbor cluster.

1. Navigate to the Harbor home address `https://core.10.10.10.100.xip.io` and login Harbor with the root user `admin` and the password you provided in the deployment manifest above.

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
