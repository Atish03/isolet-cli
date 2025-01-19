import os, time
import kubernetes_client

def wait_until_copying(resource_dir, lock_file, check_interval=2):
    print("Waiting for resources to be copied...")

    while True:
        if os.path.exists(lock_file):
            time.sleep(check_interval)
            continue

        if not os.path.exists(resource_dir):
            time.sleep(check_interval)
            continue

        return True

if __name__ == "__main__":
    source_dir = "/chall/"
    lock_file = "/tmp/resources.lock"

    wait_until_copying(source_dir, lock_file)

    print("Starting process...")
    # kcli = kubernetes_client.KubeClient()
    # print(kcli.get_registry_auth("https://index.docker.io/v1/"))
    print(os.environ)
    print("Done!")