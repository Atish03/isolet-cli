import os
import subprocess, json
from pathlib import Path

class ContainerImage():
    def __init__(self) -> None:
        self.chall_type = os.environ.get("CHALL_TYPE")
        self.registry = os.environ.get("IMAGE_REGISTRY")
    
    def __build_image(self, image_name: str) -> None:
        try:
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
        try:
            push_command = []
            auth_file = Path("/docker/.dockerconfigjson")
            if auth_file.is_file():
                push_command = [
                    "podman", "push",
                    "--authfile", "/docker/.dockerconfigjson",
                    f"{self.registry}/{image_name}:latest"
                ]
            else:
                push_command = [
                    "podman", "push",
                    f"{self.registry}/{image_name}:latest"
                ]

            result = subprocess.run(push_command, check=True, capture_output=True, text=True)
            print(f"Image pushed successfully: {image_name}")
            print(result.stdout)


        except subprocess.CalledProcessError as e:
            print(f"Failed to push image: {e}")
            print(f"Error output: {e.stderr}")
            exit(1)

    def build_and_push_images(self) -> list:
        try:
            config = None
            
            images_list = []
            
            with open("/config/config.json", "r") as f:
                config = json.loads(f.read())

            if self.chall_type != "static":
                images = config.get("docker_changed")
                for image_name in images:
                    self.__build_image(image_name)
                    self.__push_image(image_name)
                    images_list.append(f"{self.registry}/{image_name}:latest")

            print("Done!")
            return images_list
        except Exception as e:
            print(e)
            exit(1)
