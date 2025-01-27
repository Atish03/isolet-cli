import os, time
import kubernetes_client
import subprocess
import json
import container_image, database

MAX_WAIT = 150 # approx 5 mins

def wait_until_copying(resource_dir, lock_file, check_interval=2):
    print("Waiting for resources to be copied...")
    wait = 0
    while True:
        wait += 1
        if wait > MAX_WAIT:
            print("Waited too long for resources to be copied!")
            exit(1)
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

    os.chdir(source_dir)
    
    try:
        image_worker = container_image.ContainerImage()
        image_worker.build_and_push_images()
        
        db = database.Database()
        db.update_all()
        db.close()
    except Exception as e:
        print(e)
        exit(1)