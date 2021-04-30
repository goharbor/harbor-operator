# Performance

This document includes the performance of the pull/push for the Harbor deployed by the Harbor operator.

## Environment

The Harbor and Harbor operator are deployed in a Kubernetes cluster which has 3 control nodes and 3 worker nodes. Each node has 8 CPU which mode name is `Intel(R) Xeon(R) Gold 5218 CPU @ 2.30GHz` and 16GiB memory.

Here are the network benchmark for CNI using [knb](https://github.com/InfraBuilder/k8s-bench-suite) and disk benchmark for PV using [dbench](https://github.com/leeliu/dbench).

### Benchmark of network

```bash
=========================================================
 Benchmark Results
=========================================================
 Name            : knb-1763415
 Date            : 2021-04-29 01:59:31 UTC
 Generator       : knb
 Version         : 1.5.0
 Server          : harbor-wl-1-md-0-977fbcc-hlmm8
 Client          : harbor-wl-1-md-0-977fbcc-f9jhn
 UDP Socket size : auto
=========================================================
  Discovered CPU         : Intel(R) Xeon(R) Gold 5218 CPU @ 2.30GHz
  Discovered Kernel      : 5.4.0-66-generic
  Discovered k8s version : v1.18.16+vmware.1
  Discovered MTU         : 1450
  Idle :
    bandwidth = 0 Mbit/s
    client cpu =
    server cpu =
    client ram =  MB
    server ram =  MB
  Pod to pod :
    TCP :
      bandwidth = 7085 Mbit/s
      client cpu =
      server cpu =
      client ram =  MB
      server ram =  MB
    UDP :
      bandwidth = 1388 Mbit/s
      client cpu =
      server cpu =
      client ram =  MB
      server ram =  MB
  Pod to Service :
    TCP :
      bandwidth = 6909 Mbit/s
      client cpu =
      server cpu =
      client ram =  MB
      server ram =  MB
    UDP :
      bandwidth = 1467 Mbit/s
      client cpu =
      server cpu =
      client ram =  MB
      server ram =  MB
=========================================================
2021-04-29 02:01:13 [INFO] Cleaning kubernetes resources ...
```

### Benchmark of disk

```bash
==================
= Dbench Summary =
==================
Random Read/Write IOPS: 51.9k/10.1k. BW: 310MiB/s / 347MiB/s
Average Latency (usec) Read/Write: 355.22/1057.06
Sequential Read/Write: 403MiB/s / 369MiB/s
Mixed Random Read/Write IOPS: 20.0k/6950
```

## Tests

[k6](https://github.com/heww/xk6-harbor) is the testing tool to test the performance of the pull/push. Here is the scripts for the pull/push tests.

### Script for the pull test

```js
// Pull artifact which size is 10 MiB

import harbor from 'k6/x/harbor'
import { ContentStore } from 'k6/x/harbor'
import { Rate } from 'k6/metrics'

const missing = Object()

function getEnv(env, def = missing) {
    const value = __ENV[env] ? __ENV[env] : def
    if (value === missing) {
        throw (`${env} envirument is required`)
    }

    return value
}

const teardownResources = getEnv('TEARDOWN_RESOURCES', 'true') === 'true'

const artifactSize = getEnv('ARTIFACT_SIZE', '10 MiB')

const store = new ContentStore('artifact-pull')

export let successRate = new Rate('success')

export let options = {
    noUsageReport: true,
    vus: 500,
    iterations: 1000,
    thresholds: {
        'iteration_duration{scenario:default}': [
            `max>=0`,
        ],
        'iteration_duration{group:::setup}': [`max>=0`],
        'iteration_duration{group:::teardown}': [`max>=0`]
    }
};

export function setup() {
    harbor.initialize({
        scheme: getEnv('HARBOR_SCHEME', 'https'),
        host: getEnv('HARBOR_HOST'),
        username: getEnv('HARBOR_USERNAME', 'admin'),
        password: getEnv('HARBOR_PASSWORD', 'Harbor12345'),
        insecure: true,
    })

    const projectName = `project-${Date.now()}`
    try {
        harbor.createProject({ projectName })
    } catch (e) {
        console.log(e)
        errorRate.add(true)
    }

    harbor.push({
        ref: `${projectName}/benchmark:latest`,
        store,
        blobs: [store.generate(artifactSize)],
    })

    return {
        projectName,
    }
}

export default function ({ projectName }) {
    try {
        harbor.pull(`${projectName}/benchmark:latest`)
        successRate.add(true)
    } catch (e) {
        successRate.add(false)
        console.log(e)
    }
}

export function teardown({ projectName }) {
    store.free()

    if (teardownResources) {
        try {
            harbor.deleteProject(projectName, true)
        } catch (e) {
            console.log(e)
        }
    }
}
```

### Script for push test

```js
// Push artifact which size is 10 MiB

import counter from "k6/x/counter"
import harbor from 'k6/x/harbor'
import { ContentStore } from 'k6/x/harbor'
import { Rate } from 'k6/metrics'

const missing = Object()

function getEnv(env, def = missing) {
    const value = __ENV[env] ? __ENV[env] : def
    if (value === missing) {
        throw (`${env} envirument is required`)
    }

    return value
}

function getEnvInt(env, def = missing) {
    return parseInt(getEnv(env, def), 10)
}

const teardownResources = getEnv('TEARDOWN_RESOURCES', 'true') === 'true'

const artifactSize = getEnv('ARTIFACT_SIZE', '10 MiB')

const store = new ContentStore('artifact-push')

export let successRate = new Rate('success')

export let options = {
    setupTimeout: '2h',
    teardownTimeout: '1h',
    duration: '24h',
    noUsageReport: true,
    vus: 500,
    iterations: 1000,
    thresholds: {
        'iteration_duration{scenario:default}': [
            `max>=0`,
        ],
        'iteration_duration{group:::setup}': [`max>=0`],
        'iteration_duration{group:::teardown}': [`max>=0`]
    }
};

export function setup() {
    harbor.initialize({
        scheme: getEnv('HARBOR_SCHEME', 'https'),
        host: getEnv('HARBOR_HOST'),
        username: getEnv('HARBOR_USERNAME', 'admin'),
        password: getEnv('HARBOR_PASSWORD', 'Harbor12345'),
        insecure: true,
    })

    const now = Date.now()
    const projectsCount = getEnvInt('PROJECTS_COUNT', `${options.vus}`)

    const projectNames = []
    for (let i = 0; i < projectsCount; i++) {
        const projectName = `project-${now}-${i}`
        try {
            harbor.createProject({ projectName })
            projectNames.push(projectName)
        } catch (e) {
            console.log(e)
        }
    }

    const contents = store.generateMany(artifactSize, options.iterations)

    return {
        projectNames,
        contents
    }
}

export default function ({ projectNames, contents }) {
    const i = counter.up() - 1

    const projectName = projectNames[i % projectNames.length]
    const blob = contents[i % contents.length]

    try {
        const repositoryName = getEnv('REPOSITORY_NAME', `repository-${Date.now()}-${i}`)

        harbor.push({
            ref: `${projectName}/${repositoryName}:tag-${i}`,
            store,
            blobs: [blob],
        })

        successRate.add(true)
    } catch (e) {
        console.log(e)
        successRate.add(false)
    }
}

export function teardown({ projectNames }) {
    store.free()

    if (teardownResources) {
        for (const projectName of projectNames) {
            try {
                harbor.deleteProject(projectName, true)
            } catch (e) {
                console.log(`failed to delete project ${projectName}, error: ${e}`)
            }
        }
    } else {
        for (const projectName of projectNames) {
            console.log(`project ${projectName} keeped`)
        }
    }
}
```

## Manifests and results

Two different harbor clusters will be tested for the performance, one uses minio storage, another one uses filesystem storage.

### Manifests

#### Manifest of minio as storage

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
    - core.harbor.domain
    - notary.harbor.domain
    - minio.harbor.domain
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
  logLevel: debug
  imageSource:
    repository: ghcr.io/goharbor
  harborAdminPasswordRef: admin-core-secret
  externalURL: https://core.harbor.domain
  expose:
    core:
      ingress:
        host: core.harbor.domain
      tls:
        certificateRef: sample-public-certificate
    notary:
      ingress:
        host: notary.harbor.domain
      tls:
        certificateRef: sample-public-certificate
  internalTLS:
    enabled: true
  portal: {}
  registry:
    replicas: 3
    metrics:
      enabled: true
  core:
    replicas: 3
    tokenIssuer:
      name: selfsigned-issuer
      kind: Issuer
    metrics:
      enabled: false
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
      storage: 20Gi
      replicas: 1
      resources:
        limits:
          cpu: 5000m
          memory: 2500Mi
        requests:
          cpu: 1000m
          memory: 2500Mi
  inClusterStorage:
    kind: MinIO
    minIOSpec:
      replicas: 2
      secretRef: minio-access-secret
      redirect:
        enable: true
        expose:
          ingress:
            host: minio.harbor.domain
          tls:
            certificateRef: sample-public-certificate
      volumesPerServer: 2
      volumeClaimTemplate:
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 100Gi
  inClusterCache:
    kind: Redis
    redisSpec:
      server:
        replicas: 1
      sentinel:
        replicas: 1

```

#### Manifest of filesystem as storage

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
    - core.harbor.domain
    - notary.harbor.domain
  issuerRef:
    name: selfsigned-issuer
    kind: Issuer
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sample-harbor-registry
  namespace: cluster-sample-ns
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 100Gi
---
# Full stack Harbor
apiVersion: goharbor.io/v1alpha3
kind: HarborCluster
metadata:
  name: harborcluster-sample
  namespace: cluster-sample-ns
spec:
  logLevel: debug
  imageSource:
    repository: ghcr.io/goharbor
  harborAdminPasswordRef: admin-core-secret
  externalURL: https://core.harbor.domain
  expose:
    core:
      ingress:
        host: core.harbor.domain
      tls:
        certificateRef: sample-public-certificate
    notary:
      ingress:
        host: notary.harbor.domain
      tls:
        certificateRef: sample-public-certificate
  internalTLS:
    enabled: true
  portal: {}
  registry:
    replicas: 1
    metrics:
      enabled: true
  core:
    replicas: 3
    tokenIssuer:
      name: selfsigned-issuer
      kind: Issuer
    metrics:
      enabled: false
  chartmuseum: {}
  exporter: {}
  trivy:
    skipUpdate: false
    storage: {}
  notary:
    migrationEnabled: true
  imageChartStorage:
    filesystem:
      registryPersistentVolume:
        claimName: sample-harbor-registry
  inClusterDatabase:
    kind: PostgresSQL
    postgresSqlSpec:
      storage: 20Gi
      replicas: 1
      resources:
        limits:
          cpu: 5000m
          memory: 2500Mi
        requests:
          cpu: 1000m
          memory: 2500Mi
  inClusterCache:
    kind: Redis
    redisSpec:
      server:
        replicas: 1
      sentinel:
        replicas: 1

```

### Results

#### Result of minio as storage

| Test | Avg    | Min      | Med    | Max    | P90    | P95    | VUS  | Iterations |
| ---- | ------ | -------- | ------ | ------ | ------ | ------ | ---- | ---------- |
| Pull | 32.81s | 876.88ms | 31.19s | 1m25s  | 1m6s   | 1m14s  | 200  | 400        |
| Push | 29.66s | 15.09s   | 26.39s | 55.47s | 44.45s | 45.47s | 200  | 400        |

#### Result of filesystem as storage

| Test | Avg    | Min      | Med    | Max   | P90    | P95    | VUS  | Iterations |
| ---- | ------ | -------- | ------ | ----- | ------ | ------ | ---- | ---------- |
| Pull | 32.34s | 838.94ms | 27.81s | 1m24s | 1m6s   | 1m14s  | 200  | 400        |
| Push | 20.43s | 10.62s   | 20.33s | 35s   | 24.92s | 28.18s | 200  | 400        |
