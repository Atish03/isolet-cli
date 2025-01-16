import docker
import json
import os
import kubernetes_client

def build_and_push_image(image_name, tag, dockerfile_path, repository_url):
    client = docker.from_env()

    try:
        auth_config = kubernetes_client.get_registry_auth()
        if auth_config == None:
            raise ValueError("No secret for dockerconfig found")

        full_image_name = f"{repository_url}/{image_name}:{tag}"
        print(f"Building image {full_image_name}...")
        image, logs = client.images.build(path=dockerfile_path, tag=full_image_name)

        for log in logs:
            if "stream" in log:
                print(log["stream"], end="")

        print(f"Image {full_image_name} built successfully.")

        print(f"Pushing image {full_image_name} to {repository_url}...")
        for line in client.images.push(full_image_name, stream=True, auth_config=auth_config):
            print(json.loads(line.decode("utf-8")).get("status", ""), end="\r")

        print(f"Image {full_image_name} pushed successfully.")

    except Exception as e:
        print(f"Error: {e}")
    finally:
        client.close()
