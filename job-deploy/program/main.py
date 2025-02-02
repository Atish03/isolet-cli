import database, patch

db = database.Database()
p = patch.Patch()

challs = ["SQL Injection Lab"]

res = db.get_deploy_meta(challs)

for chall in res:
    if chall["deployment"] != "http":
        p.add_port_to_service(f"{chall['subd']}-port", chall["port"])
        p.add_entrypoint(chall["subd"], f":{chall['port']}")
    
p.commit()