import requests
import jwt
import os, json
from urllib.parse import urljoin

class Resources:
    def __init__(self):
        self.data = {
            "email": "jithendranadh15@gmail.com",
            "exp": 1820448995,
            "rank": 3,
            "teamid": -1,
            "userid": 1
        }
        self.secret = os.environ.get("ADMIN_SECRET")
        self.public_url = os.environ.get("PUBLIC_URL")
    
    def __upload(self, file_path: str):
        URL = "http://fileserver-svc.platform.svc.cluster.local/admin/upload"
        TOKEN = jwt.encode(self.data, self.secret, algorithm="HS256")
        
        cookies = {
            "token": TOKEN,
        }
        
        files = {
            "file": open(file_path, 'rb')
        }
        
        res = requests.post(URL, cookies=cookies, files=files)
        
        if res.status_code >= 200 or res.status_code <= 300:
            return os.path.join("https://", self.public_url, res.text.strip("/"))
        else:
            print("Not uploaded:", res.content.decode())
            return None
    
    def upload(self) -> list:
        public_urls = []
        
        # config = None
            
        # with open("/config/config.json", "r") as f:
        #     config = json.loads(f.read())
            
        # files = config.get("res_changed")
        
        if os.path.exists("/chall/resources/"):
            files = os.listdir("/chall/resources/")
            
            for f in files:
                file_path = os.path.join("/chall/resources/", f)
                upload_path = self.__upload(file_path)
                if upload_path:
                    public_urls.append(upload_path)
        
        return public_urls