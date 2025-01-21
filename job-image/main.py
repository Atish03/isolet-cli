import os, time
import kubernetes_client
import subprocess

MAX_WAIT = 150 # approx 5 mins

def wait_until_copying(resource_dir, lock_file, check_interval=2):
    print("Waiting for resources to be copied...")
    wait = 0
    while True:
        wait += 1
        if wait > MAX_WAIT:
            print("Waited too long for resources to be copied!")
            exit(0)
        if os.path.exists(lock_file):
            time.sleep(check_interval)
            continue

        if not os.path.exists(resource_dir):
            time.sleep(check_interval)
            continue

        return True

def build_image(tag: str) -> None:
    try:
        build_command = [
            "buildah", "bud",
            "--file", "/chall/Dockerfile",
            "--tag", image_name
        ]
        
        print(f"Running command: {' '.join(build_command)}")
        
        result = subprocess.run(build_command, check=True, capture_output=True, text=True)
        print(f"Image built successfully: {image_name}")
        print(result.stdout)
    
    except subprocess.CalledProcessError as e:
        print(f"Failed to build image: {e}")
        print(f"Error output: {e.stderr}")
        exit(0)

def push_image(registry: str, image_name: str) -> None:
    client = kubernetes_client.KubeClient()
    auth = client.get_registry_auth(registry)

    cred = f"{auth.get('username')}:{auth.get('password')}"

    try:
        push_command = [
            "buildah", "push",
            "--creds", cred,
            image_name
        ]

        print(f"Running command: {' '.join(push_command)}")
        
        result = subprocess.run(push_command, check=True, capture_output=True, text=True)
        print(f"Image pushed successfully: {image_name}")
        print(result.stdout)


    except subprocess.CalledProcessError as e:
        print(f"Failed to push image: {e}")
        print(f"Error output: {e.stderr}")
        exit(0)


if __name__ == "__main__":
    source_dir = "/chall/"
    lock_file = "/tmp/resources.lock"

    wait_until_copying(source_dir, lock_file)

    os.chdir("/chall")

    # LOGIC
    image_name = os.environ.get("CHALL_IMAGE_NAME")
    registry_name = os.environ.get("IMAGE_REGISTRY")
    chall_type = os.environ.get("CHALL_TYPE")

    if chall_type != "static":
        build_image(image_name)
        push_image(registry_name, image_name)

    print("Done!")
