from kubernetes import client, config
import yaml
import datetime

class Patch:
    def __init__(self):
        config.load_incluster_config()
        self.v1 = client.CoreV1Api()
        self.api = client.AppsV1Api()
        
        self.TRAEFIK_CONFIG     = "traefik-config"
        self.TRAEFIK_SERVICE    = "traefik-svc"
        self.TRAEFIK_NAMESPACE  = "traefik"
        self.TRAEFIK_DEPLOYMENT = "traefik-deployment"
        
        # ======================= ENTRYPOINTS CONFIGMAP =======================
        self.configmap = self.v1.read_namespaced_config_map(name=self.TRAEFIK_CONFIG, namespace=self.TRAEFIK_NAMESPACE)
        self.config_data = yaml.safe_load(self.configmap.data.get("traefik.yaml", ""))
        
        # ======================= LOAD BALANCER SERVICE =======================
        self.service = self.v1.read_namespaced_service(name=self.TRAEFIK_SERVICE, namespace=self.TRAEFIK_NAMESPACE)
        
        # ======================== TRAEFIK DEPLOYMENT =========================
        self.deployment = self.api.read_namespaced_deployment(name=self.TRAEFIK_DEPLOYMENT, namespace=self.TRAEFIK_NAMESPACE)
        

    def add_entrypoint(self, entrypoint: str, address: str) -> None:        
        self.config_data["entryPoints"][entrypoint] = {
            "address": address
        }
        
    def add_port_to_service(self, port_name: str, port: int) -> None:
        new_port = {
            "name": port_name,
            "port": port,
            "target_port": port,
            "protocol": "TCP"
        }
        
        if not any(p.port == new_port["port"] for p in self.service.spec.ports):
            self.service.spec.ports.append(client.V1ServicePort(**new_port))
        else:
            print(f"Port {new_port['port']} already exists in {self.TRAEFIK_SERVICE}.")
    
    def commit(self):
        self.configmap.data["traefik.yaml"] = yaml.dump(self.config_data)
        self.v1.replace_namespaced_config_map(name=self.TRAEFIK_CONFIG, namespace=self.TRAEFIK_NAMESPACE, body=self.configmap)
        
        self.v1.patch_namespaced_service(name=self.TRAEFIK_SERVICE, namespace=self.TRAEFIK_NAMESPACE, body=self.service)
        
        if not self.deployment.spec.template.metadata.annotations:
            self.deployment.spec.template.metadata.annotations = {}
            
        self.deployment.spec.template.metadata.annotations["kubectl.kubernetes.io/restartedAt"] = datetime.datetime.utcnow().isoformat()
        self.api.patch_namespaced_deployment(name=self.TRAEFIK_DEPLOYMENT, namespace=self.TRAEFIK_NAMESPACE, body=self.deployment)