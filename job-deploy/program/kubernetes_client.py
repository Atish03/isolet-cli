from kubernetes import client, config, utils
import base64, json
import os
import yaml
from patch import Patch

class KubeClient():
    def __init__(self, incluster: bool=True) -> None:
        if incluster:
            config.load_incluster_config()
        else:
            config.load_kube_config()
            
        self.v1 = client.CoreV1Api()
        self.api = client.ApiClient()
        self.custom_api = client.CustomObjectsApi()
        
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
            db_config["host"] = base64.b64decode(decoded_data["POSTGRES_HOST"]).decode()
        except Exception as e:
            print(f"Error reading Secret: {e}")
            exit(1)
            return None
        
        return db_config
    
    def __get_deployment(self, subd: str) -> str:
        try:
            cm = self.v1.read_namespaced_config_map(name=f"{subd}-cm", namespace="store")
            return cm.data["deployment.yaml"]
        except Exception as e:
            print("cannot read deployment.yaml,", e)
            exit(1)
        
    def apply(self, subd: str, p: Patch) -> None:
        try:    
            yaml_objects = list(yaml.safe_load_all(self.__get_deployment(subd)))
            
            group = "traefik.io"
            version = "v1alpha1"
            namespace = "dynamic"

            for yaml_obj in yaml_objects:
                if yaml_obj["kind"] == "IngressRoute" or yaml_obj["kind"] == "IngressRouteTCP":
                    if yaml_obj["kind"] == "IngressRoute":
                        plural = "ingressroutes"
                    else:
                        plural = "ingressroutetcps"
                    self.custom_api.create_namespaced_custom_object(
                        group=group,
                        version=version,
                        namespace=namespace,
                        plural=plural,
                        body=yaml_objects[-1],
                    )
                else:
                    utils.create_from_dict(self.api, yaml_obj, namespace=namespace)
        
        except utils.FailToCreateError as e:
            exceptions = e.api_exceptions
            
            for exception in exceptions:
                body = json.loads(exception.body)
                if body.get("code") == 409:
                    print(f"{body.get('message')}, please undeploy first.")
                else:
                    print("error when applying:", e)
                    
            exit(1)