import os, time
import kubernetes_client
import subprocess
import json
import container_image, database, resources
import traceback

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
    
    try:
        image_worker = container_image.ContainerImage()
        phoros_worker = resources.Resources()
        db = database.Database()
        
        image_names = image_worker.build_and_push_images()
        urls = phoros_worker.upload()
        chall_id = db.update_all()
        
        if len(image_names) == 1:
            db.update_images_table(chall_id, image_names[0])
        
        db.update_links(chall_id, urls)
         
        db.commit()
        db.close()
    except Exception as e:
        print(traceback.format_exc())
        exit(1)