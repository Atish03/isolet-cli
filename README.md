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
> The `chall ls` command will help you list challenges in a particular directory. The directory can be the challenge directory or the parent directory of the challenge directories.

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

deployment_type: http
deployment_port: 80
cpu (optional): 15
mem (optional): 32
```

## Deployment
The `deploy` command is used to deploy dynamic challenge and expose it using a load-balancer and traefik. The usage is the exact same as  `load` command.
Similarly one can take down a deployment using `undeploy` command

> [!IMPORTANT]
> Try to work on a challenge on a single device, since syncing might cause issues, make sure that the cli tool is not used to edit challenges after the event is started, all the challenges must be managed using admin panel (except deploying dynamic challenges) since mismatch of challenge metadata may cause issues 

## Custom deployment
Typically a dynamic challenge must use a single docker container which will be run and specified port will be exposed, but there might be a case where author wants to use multiple containers.

One can use `deployment.yaml` to define their deployment and run `chall deploy`

Following is a template for `deployment.yaml`:
```yaml
# DO NOT CHANGE ANYTHING, JUST ADD CONTAINERS, CHANGE PORTS AND EDIT INGRESS SPECS
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Subd}}
  namespace: dynamic
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.Subd}}
  template:
    metadata:
      labels:
        app: {{.Subd}}
    spec:
      imagePullSecrets:
      - name: dynamic-registry-secret
      containers:
      - name: "container-1"
        image: <registry/image-1>
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 72m
            memory: 128Mi
        ports:
        - containerPort: 22 # The port of container you want to expose
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Subd}}-svc
  namespace: dynamic
annotations:
  "traefik.ingress.kubernetes.io/router.entrypoints": {{.Subd}}
spec:
  selector:
    app: {{.Subd}}
  ports:
  - port: 8008
    targetPort: 22 # Change this to containerPort
---
apiVersion: traefik.io/v1alpha1
kind: IngressRouteTCP
metadata:
  name: {{.Subd}}-ingress
  namespace: dynamic
spec:
  entryPoints: # The entrypoints for http challenge will be web and websecure
  - {{.Subd}}
  tls: # Only for http challenges
    certResolver: default
  routes:
  - match: HostSNI(`*`) # Make this Host(`subdomain.domain`) for http challenges
    kind: Rule
    middlewares: # Only for http challenges
      - name: secure-headers
    services:
    - name: {{.Subd}}-svc
      port: 8008
```

## Installation
```
./configure --prefix=<install dir>
make install

# uninstall isolet
make uninstall
```
