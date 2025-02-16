import database, patch
import kubernetes_client
import json

kube_client = kubernetes_client.KubeClient()

db = database.Database()
p = patch.Patch()

challs = []

with open("/config/challs.json", "r") as f:
    challs = json.loads(f.read()).get("challs", [])
    
print("Challenges to deploy: \n", [x["chall_name"] for x in challs])

for chall in challs:    
    if chall["deployment_config"]["type"] != "http":
        p.add_port_to_service(f"{chall['deployment_config']['subd']}-port", chall["deployment_config"]["port"])
        p.add_entrypoint(chall["deployment_config"]["subd"], f":{chall['deployment_config']['port']}")
        
    kube_client.apply(chall["deployment_config"]["subd"], p)
    
p.commit()