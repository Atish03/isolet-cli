import os
import kubernetes_client
import subprocess, json

class ContainerImage():
    def __init__(self) -> None:
        self.registry_url = os.environ.get("IMAGE_REGISTRY_URL")
        self.chall_type = os.environ.get("CHALL_TYPE")
        self.registry = os.environ.get("IMAGE_REGISTRY")
        client = kubernetes_client.KubeClient()
        self.auth = client.get_registry_auth(self.registry_url)
    
    def __build_image(self, image_name: str) -> None:
        try:
            os.chdir(os.path.join("/chall/Dockerfiles", image_name))
            build_command = [
                "podman", "build",
                "--file", os.path.join("/chall/Dockerfiles", image_name, "Dockerfile"),
                "--tag", f"{self.registry}/{image_name}:latest"
            ]
            
            result = subprocess.run(build_command, check=True, capture_output=True, text=True)
            print(f"Image built successfully: {image_name}")
            print(result.stdout)
        
        except subprocess.CalledProcessError as e:
            print(f"Failed to build image: {e}")
            print(f"Error output: {e.stderr}")
            exit(1)

    def __push_image(self, image_name: str) -> None:
        cred = f"{self.auth.get('username')}:{self.auth.get('password')}"

        try:
            push_command = [
                "podman", "push",
                "--creds", cred,
                f"{self.registry}/{image_name}:latest"
            ]

            result = subprocess.run(push_command, check=True, capture_output=True, text=True)
            print(f"Image pushed successfully: {image_name}")
            print(result.stdout)


        except subprocess.CalledProcessError as e:
            print(f"Failed to push image: {e}")
            print(f"Error output: {e.stderr}")
            exit(1)

    def build_and_push_images(self):
        try:
            config = None
            
            with open("/config/config.json", "r") as f:
                config = json.loads(f.read())

            if self.chall_type != "static":
                images = config.get("docker_changed")
                for image_name in images:
                    self.__build_image(image_name)
                    self.__push_image(image_name)

            print("Done!")
        except Exception as e:
            print(e)
            exit(1)
