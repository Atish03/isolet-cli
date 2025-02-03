# Isolet CLI

`isolet-cli` is a command line tool for [isolet](https://github.com/thealpha16/isolet) to manage challenges on isolet platform. 

> Use `--help` flag to know more about a particular command.

## Challenge directory
A challenge is a directory with the following structure

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
├── deployment.yaml
└── resources
    └── chall.zip
```

> [!NOTE]
>the `chall ls` command will help you list challenges in a particular directory. The directory can be the challenge directory or the parent directory of the challenge directories.

## chall.yaml
`chall.yaml` file must be of following format

```
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
