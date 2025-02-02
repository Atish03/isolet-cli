from kubernetes import client, config
import base64, json

class KubeClient():
    def __init__(self, incluster: bool=True) -> None:
        if incluster:
            config.load_incluster_config()
        else:
            config.load_kube_config()
            
        self.v1 = client.CoreV1Api()
        
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