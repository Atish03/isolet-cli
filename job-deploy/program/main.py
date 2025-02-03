import database, patch
import kubernetes_client
import json

kube_client = kubernetes_client.KubeClient()

db = database.Database()
p = patch.Patch()

challs = []

with open("/config/challs.json", "r") as f:
    challs = json.loads(f.read()).get("challs", [])
    
print("Challenges to deploy: \n", challs)

res = db.get_deploy_meta(challs)

for chall in res:
    port = 80
    
    if chall["deployment"] == "nc":
        port = 6969
    elif chall["deployment"] == "ssh":
        port = 22
        
    if chall["deployment"] != "http":
        p.add_port_to_service(f"{chall['subd']}-port", chall["port"])
        p.add_entrypoint(chall["subd"], f":{chall['port']}")
        
    kube_client.expose_chall(chall["subd"], chall["image"], port, chall["private-registry"])
    
p.commit()