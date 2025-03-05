from kubernetes import client, config
import yaml, json
import os

class Deployment:
    def __init__(self, incluster: bool=True):
        if incluster:
            config.load_incluster_config()
        else:
            config.load_kube_config()
        
        self.config = None
        with open("/config/config.json", "r") as f:
            self.config = json.loads(f.read())
            
        self.public_url = os.environ.get("PUBLIC_URL")
        self.core_api = client.CoreV1Api()
        self.api = client.ApiClient()
    
    def __get_deployment_yaml(self, subd: str, image: str, port: int, private: bool, resources: dict, secret: str, namespace: str) -> dict:
        try:
            deployment = client.V1Deployment(
                api_version="apps/v1",
                kind="Deployment",
                metadata=client.V1ObjectMeta(
                    name=subd,
                    namespace=namespace,
                    labels={"app.kubernetes.io/component": "deployment", "app.kubernetes.io/name": subd, "app.kubernetes.io/part-of": "challenges"}
                ),
                spec=client.V1DeploymentSpec(
                    replicas=1,
                    selector=client.V1LabelSelector(
                        match_labels={"app.kubernetes.io/component": "deployment", "app.kubernetes.io/name": subd, "app.kubernetes.io/part-of": "challenges"}
                    ),
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels={"app.kubernetes.io/component": "deployment", "app.kubernetes.io/name": subd, "app.kubernetes.io/part-of": "challenges"}),
                        spec=client.V1PodSpec(
                            containers=[
                                client.V1Container(
                                    name=subd,
                                    image=image,
                                    resources=client.V1ResourceRequirements(
                                        limits={
                                            "cpu": resources["cpu_limit"],
                                            "memory": resources["mem_limit"],
                                        },
                                    ),
                                    ports=[client.V1ContainerPort(container_port=port)]
                                )
                            ]
                        )
                    )
                )
            )
            
            if private:
                deployment.spec.template.spec.image_pull_secrets = [client.V1LocalObjectReference(name=secret)]
                
            return self.api.sanitize_for_serialization(deployment)
            
        except Exception as e:
            print(e)
            exit(1)
            
    def __get_svc_yaml(self, subd: str, port: int, namespace: str) -> dict:
        try:
            service = client.V1Service(
                api_version="v1",
                kind="Service",
                metadata=client.V1ObjectMeta(
                    name=f"{subd}-svc",
                    namespace=namespace,
                    annotations={
                        "traefik.ingress.kubernetes.io/router.entrypoints": subd
                    },
                    labels={
                        "app.kubernetes.io/component": "service",
                        "app.kubernetes.io/part-of": "challenges",
                        "app.kubernetes.io/name": subd
                    }
                ),
                spec=client.V1ServiceSpec(
                    selector={"app.kubernetes.io/name": subd},
                    ports=[
                        client.V1ServicePort(
                            protocol="TCP",
                            port=8008,
                            target_port=port
                        )
                    ],
                    type="ClusterIP"
                )
            )
            
            return self.api.sanitize_for_serialization(service)
            
        except Exception as e:
            print(e)
            exit(1)
            
    def __get_ingress_yaml(self, subd: str, host: str, dep_type: str, namespace: str) -> dict:
        try:
            entrypoints = [subd]
            ingress_type = "IngressRouteTCP"
            
            if dep_type == "http":
                entrypoints = ["web", "websecure"]
                ingress_type = "IngressRoute"
            
            ingress_route_tcp = {
                "apiVersion": "traefik.io/v1alpha1",
                "kind": ingress_type,
                "metadata": {
                    "name": f"{subd}-ingress",
                    "namespace": namespace,
                    "labels": {
                        "app.kubernetes.io/component": "ingress",
                        "app.kubernetes.io/part-of": "challenges",
                        "app.kubernetes.io/name": subd
                    }
                },
                "spec": {
                    "entryPoints": entrypoints,
                    "routes": [
                        {
                            "match": host,
                            "kind": "Rule",
                            "services": [
                                {
                                    "name": f"{subd}-svc",
                                    "port": 8008
                                }
                            ]
                        }
                    ]
                }
            }
            
            if dep_type == "http":
                ingress_route_tcp["spec"]["tls"] = {"secretName": "challenge-certs"}
            
            return self.api.sanitize_for_serialization(ingress_route_tcp)
        
        except Exception as e:
            print(e)
            exit(1)
            
    def create(self, image: list):
        chall_type = os.environ.get("CHALL_TYPE")
        namespace = "isolet"
        
        if chall_type == "dynamic":
            namespace = "dynamic"
        
        if chall_type != "static":
            deployment_config = self.config["deployment_config"]
            
            if not deployment_config["custom_deploy"]["custom"]:
                port = 80
                if deployment_config["type"] == "ssh":
                    port = 22
                if deployment_config["type"] == "nc":
                    port = 6969
                    
                host = "HostSNI(`*`)"
                if port == 80:
                    host = f"Host(`{deployment_config['subd']}.ctf.{self.public_url}`)"
                
                dep = self.__get_deployment_yaml(deployment_config["subd"], image[0], port, deployment_config["registry"]["private"], deployment_config["resources"], deployment_config["registry"]["secret"], namespace)
                svc = self.__get_svc_yaml(deployment_config["subd"], port, namespace)
                ing = self.__get_ingress_yaml(deployment_config["subd"], host, deployment_config["type"], namespace)
                
                deployment_yaml = yaml.safe_dump_all([dep, svc, ing])
                
                body = client.V1ConfigMap(
                    api_version="v1",
                    kind="ConfigMap",
                    metadata=client.V1ObjectMeta(name=f"{deployment_config['subd']}-cm", namespace="store", labels={"app.kubernetes.io/component": "config", "app.kubernetes.io/part-of": "challenges"}),
                    data={
                        "deployment.yaml": deployment_yaml
                    }
                )
                
                try:
                    self.core_api.create_namespaced_config_map(namespace=body.metadata.namespace, body=body)
                except client.exceptions.ApiException as e:
                    if e.status == 409:
                        self.core_api.replace_namespaced_config_map(name=body.metadata.name, namespace=body.metadata.namespace, body=body)
                    else:
                        print(e)
                        exit(1)