import psycopg2, json
import psycopg2.extras
from string import Template
import kubernetes_client

class Database():
    def __init__(self) -> None:            
        self.client = kubernetes_client.KubeClient()
        db_config = self.client.get_db_config()
        
        if db_config == None:
            print("db config was not found")
            exit(1)
        
        self.conn = psycopg2.connect(**db_config)
        self.cursor = self.conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)
    
    def __get_chall_id(self, chall_name: str) -> int:
        query = """
        SELECT chall_id FROM challenges
        WHERE chall_name = %s
        """
        
        self.cursor.execute(query, (chall_name,))
        row = self.cursor.fetchone()
        
        return row["chall_id"]
    
    def get_deploy_meta(self, challs: list) -> list:
        query = """
        SELECT image, deployment, port, subd, cpu, mem
        FROM images
        WHERE chall_id = %s
        """
        
        res = []
        
        for chall in challs:
            chall_id = self.__get_chall_id(chall.get("chall_name", ""))
            self.cursor.execute(query, (chall_id,))
            
            row = dict(self.cursor.fetchone())
            row["private-registry"] = self.client.check_private()
            row["custom"] = chall.get("custom", False)
            row["yaml_string"] = chall.get("yaml_string", "")
            res.append(row)
            
        return res
    
    def close(self):
        if self.cursor:
            self.cursor.close()
        if self.conn:
            self.conn.close()