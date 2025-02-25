# Isolet CLI

`isolet-cli` is a command line tool for [isolet](https://github.com/thealpha16/isolet) to manage challenges on isolet platform. 

> Use `--help` flag to know more about a particular command.

```
.
├── Dockerfiles
│   ├── chall-image-1
│   │   ├── Dockerfile
│   │   └── docker-files
│   │       └── file.txt
│   └── image-2
│       └── Dockerfile
├── chall.yaml
├── deployment.yaml (optional) # Refer custom deployment
└── resources
    └── chall.zip
```

> [!NOTE]
> The `chall ls` command will help you list challenges in a particular directory. The directory can be the challenge directory or the parent directory of the challenge directories, (DIRECTORY NAMES MUST FOLLOW RFC 1123).

> [!CAUTION]
> When defining `Dockerfile` for dynamic challenges make sure to use port `80 for http challenges`, `22 for ssh challenges` and `6969 for nc challenges`.


## chall.yaml
`chall.yaml` file must be of following format

```yaml
chall_name: "SQL Injection Lab"
type: "dynamic"
category_name: "Web"
prompt: "Find the vulnerability in the login system to retrieve the flag."
points: 200
flag: "CTF{dynamic_sql_injection_flag}"
attempts (optional): 20 
hints (optional):
  - hint: "Start by trying simple ' OR '1'='1' payloads"
    visible: false
  - hint: "Use tools like sqlmap if stuck"
    cost: 300
author: "Bob"
visible (optional): false
tags:
  - "Web"
  - "Beginner"
links (optional):
  - "https://wiki.com/sql_injection"
files (optional):
  - "file.txt" # Must match the file name in resources directory

# Only for dynamic challenges

deployment_type: ssh
deployment_port: 2345 # Not required if deployment type is http or challenge type is on-demand
cpu (optional): 15m
mem (optional): 32Mi
```

## Deployment
The `deploy` command is used to deploy dynamic challenge and expose it using a load-balancer and traefik. The usage is the exact same as  `load` command.
Similarly one can take down a deployment using `undeploy` command

> [!IMPORTANT]
> Try to work on a challenge on a single device, since syncing might cause issues, make sure that the cli tool is not used to edit challenges after the event is started, all the challenges must be managed using admin panel.

## Custom deployment
Typically a dynamic/on-demand challenge must use a single docker container which will be run and specified port will be exposed, but there might be a case where author wants to use multiple containers.

One can use `deployment.yaml` to define their deployment and run `chall load`

Following is a template for `deployment.yaml`:
```yaml
# DO NOT CHANGE ANYTHING EXCEPT WHEREVER COMMENTS ARE PRESENT AND ADD CONTAINERS
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Subd}}
  namespace: dynamic
  labels:
    app.kubernetes.io/component: deployment
    app.kubernetes.io/name: {{.Subd}}
    app.kubernetes.io/part-of: challenges
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: deployment
      app.kubernetes.io/name: {{.Subd}}
      app.kubernetes.io/part-of: challenges
  template:
    metadata:
      labels:
        app.kubernetes.io/component: deployment
        app.kubernetes.io/name: {{.Subd}}
        app.kubernetes.io/part-of: challenges
    spec:
      containers:
      - image: {{.Registry}}/<image-dir-name> # put the name of directory your required image's Dockerfile is in
        name: container-1
        ports:
        - containerPort: 80 # 22 for ssh, 6969 for nc
        resources:
          limits:
            cpu: 300m
            memory: 256Mi
          requests:
            cpu: 100m
            memory: 64Mi
      imagePullSecrets:
      - name: challenge-registry-secret
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: {{.Subd}}
  name: {{.Subd}}-svc
  namespace: dynamic
  labels:
    app.kubernetes.io/component: service
    app.kubernetes.io/name: {{.Subd}}
    app.kubernetes.io/part-of: challenges
spec:
  ports:
  - port: 8008
    protocol: TCP
    targetPort: 80 # same as containerPort
  selector:
    app.kubernetes.io/name: {{.Subd}}
  type: ClusterIP
---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: {{.Subd}}-ingress
  namespace: dynamic
  labels:
    app.kubernetes.io/component: ingress
    app.kubernetes.io/name: {{.Subd}}
    app.kubernetes.io/part-of: challenges
spec:
  entryPoints:
  - {{.Subd}} # only for nc/ssh challs
  - web # only for http challs
  - websecure # only for http challs
  routes:
  - kind: Rule
    match: Host(`subdomain.domain.tld`) # use HostSNI(`*`) for nc/ssh challs
    services:
    - name: {{.Subd}}-svc
      port: 8008
  tls: # only for http challs
    secretName: challenge-certs
```

## Installation
```
./configure --prefix=<install dir>
make install

# uninstall isolet
make uninstall
```
