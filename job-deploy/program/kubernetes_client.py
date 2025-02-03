from kubernetes import client, config
import base64, json
import os

class KubeClient():
    def __init__(self, incluster: bool=True) -> None:
        if incluster:
            config.load_incluster_config()
        else:
            config.load_kube_config()
            
        self.v1 = client.CoreV1Api()
        self.app = client.AppsV1Api()
        self.custom_api = client.CustomObjectsApi()
        self.DOMAIN = os.environ.get("DOMAIN_NAME", "localhost")
        
    def get_db_config(self) -> dict:
        db_config = {
            'dbname': "testdb",
            'user': "begula",
            'password': "passwd",
            'host': "127.0.0.1",
            'port': 5432,
        }
        
        try:
            secret = self.v1.read_namespaced_secret(name="automation-secrets", namespace="automation")
            decoded_data = {key: value.encode('utf-8').decode('utf-8') for key, value in secret.data.items()}
            
            db_config["dbname"] = base64.b64decode(decoded_data["POSTGRES_DATABASE"]).decode()
            db_config["user"] = base64.b64decode(decoded_data["POSTGRES_USER"]).decode()
            db_config["password"] = base64.b64decode(decoded_data["POSTGRES_PASSWORD"]).decode()
            db_config["host"] = f'{base64.b64decode(decoded_data["POSTGRES_HOST"]).decode()}.platform.svc.cluster.local'
        except Exception as e:
            print(f"Error reading Secret: {e}")
            exit(1)
            return None
        
        return db_config
    
    def check_private(self) -> bool:
        try:
            self.v1.read_namespaced_secret(name="dynamic-registry-secret", namespace="automation")
            return True
        except Exception as e:
            return False
        
    def __deploy_pods(self, subd: str, image: str, port: int, private: bool) -> None:
        try:
            deployment = client.V1Deployment(
                api_version="apps/v1",
                kind="Deployment",
                metadata=client.V1ObjectMeta(
                    name=subd,
                    namespace="dynamic"
                ),
                spec=client.V1DeploymentSpec(
                    replicas=1,
                    selector=client.V1LabelSelector(
                        match_labels={"app": subd}
                    ),
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels={"app": subd}),
                        spec=client.V1PodSpec(
                            containers=[
                                client.V1Container(
                                    name=subd,
                                    image=image,
                                    ports=[client.V1ContainerPort(container_port=port)]
                                )
                            ]
                        )
                    )
                )
            )
            
            if private:
                deployment.spec.template.spec.image_pull_secrets = [client.V1LocalObjectReference(name="dynamic-registry-secret")]
                
            self.app.create_namespaced_deployment(namespace="dynamic", body=deployment)
            
        except Exception as e:
            print(e)
            
    def __start_svc(self, subd: str, port: int) -> None:
        try:
            service = client.V1Service(
                api_version="v1",
                kind="Service",
                metadata=client.V1ObjectMeta(
                    name=f"{subd}-svc",
                    namespace="dynamic",
                    annotations={
                        "traefik.ingress.kubernetes.io/router.entrypoints": subd
                    }
                ),
                spec=client.V1ServiceSpec(
                    selector={"app": subd},
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
            
            self.v1.create_namespaced_service(namespace="dynamic", body=service)
            
        except Exception as e:
            print(e)
            
    def __define_ingress(self, subd: str, url: str=None) -> None:
        try:
            routes_match = "HostSNI(`*`)"
            entrypoints = [subd]
            
            if url:
                routes_match = f"HostSNI(`{url}`)"
                entrypoints = ["web", "websecure"]
            
            ingress_route_tcp = {
                "apiVersion": "traefik.io/v1alpha1",
                "kind": "IngressRouteTCP",
                "metadata": {
                    "name": f"{subd}-ingress",
                    "namespace": "dynamic"
                },
                "spec": {
                    "entryPoints": entrypoints,
                    "routes": [
                        {
                            "match": routes_match,
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
            
            # if url:
            #     ingress_route_tcp["spec"]["routes"][0]["middlewares"] = [{"name": "secure-headers"}]
            
            group = "traefik.io"
            version = "v1alpha1"
            plural = "ingressroutetcps"
            
            self.custom_api.create_namespaced_custom_object(
                group=group,
                version=version,
                namespace="dynamic",
                plural=plural,
                body=ingress_route_tcp
            )
        
        except Exception as e:
            print(e)
            
    def expose_chall(self, subd: str, image: str, port: int, private: bool):
        url = None
        if port == 80:
            url = f"{subd}.ctf.{self.DOMAIN}"
        
        self.__deploy_pods(subd, image, port, private)
        self.__start_svc(subd, port)
        self.__define_ingress(subd, url)