from kubernetes import client, config
import base64, json

class KubeClient():
    def __init__(self, incluster: bool=True) -> None:
        if incluster:
            config.load_incluster_config()
        else:
            config.load_kube_config()

    def __get_docker_config(self, namespace: str, secret_name: str) -> str:
        try:
            v1 = client.CoreV1Api()
            secret = v1.read_namespaced_secret(name=secret_name, namespace=namespace)
            decoded_data = {key: value.encode('utf-8').decode('utf-8') for key, value in secret.data.items()}

            if not decoded_data.__contains__(".dockerconfigjson"):
                return None

            return decoded_data[".dockerconfigjson"]
        except Exception as e:
            print(f"Error reading Secret: {e}")
            return None

    def get_registry_auth(self, registry_url: str) -> dict:
        enc_config = self.__get_docker_config("automation", "build-registry-secret")
        
        if enc_config != None:
            config = json.loads(base64.b64decode(enc_config).decode())
            if "auths" not in config:
                raise ValueError("No auths found in Docker config.")

            repository_auth = config["auths"].get(registry_url)
            if not repository_auth:
                raise ValueError(f"No credentials found for repository: {registry_url}")

            auth = base64.b64decode(repository_auth["auth"]).decode().split(":")
            
            auth_config = {
                "username": auth[0],
                "password": auth[1],
            }

            return auth_config
        else:
            return None
        
    def get_db_config(self) -> dict:
        db_config = {
            'dbname': "testdb",
            'user': "begula",
            'password': "passwd",
            'host': "127.0.0.1",
            'port': 5432,
        }
        
        try:
            v1 = client.CoreV1Api()
            secret = v1.read_namespaced_secret(name="automation-secrets", namespace="automation")
            decoded_data = {key: value.encode('utf-8').decode('utf-8') for key, value in secret.data.items()}
            
            db_config["dbname"] = base64.b64decode(decoded_data["POSTGRES_DATABASE"]).decode()
            db_config["user"] = base64.b64decode(decoded_data["POSTGRES_USER"]).decode()
            db_config["password"] = base64.b64decode(decoded_data["POSTGRES_PASSWORD"]).decode()
            db_config["host"] = f'{base64.b64decode(decoded_data["POSTGRES_HOST"]).decode()}.platform.svc.cluster.local'
        except Exception as e:
            print(f"Error reading Secret: {e}")
            return None
        
        return db_config