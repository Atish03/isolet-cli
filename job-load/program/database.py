import psycopg2, json
from string import Template
import kubernetes_client

class Database():
    def __init__(self) -> None:
        self.config = None
        with open("/config/config.json", "r") as f:
            self.config = json.loads(f.read())
            
        client = kubernetes_client.KubeClient()
        db_config = client.get_db_config()
        
        if db_config == None:
            print("db config was not found")
            exit(1)
        
        self.conn = psycopg2.connect(**db_config)
        self.cursor = self.conn.cursor()
    
    def __insert_category(self) -> int:
        self.cursor.execute(self.config.get('category_query'))
        rows = self.cursor.fetchall()
        
        return rows[0][0]
    
    def __change_chall_name(self) -> None:
        query = """
        UPDATE challenges
        SET chall_name = %s
        WHERE chall_name = %s
        """
        
        old_name = self.config.get("old_name", "")
        new_name = self.config.get("new_name", "")
        if old_name != new_name:
            self.cursor.execute(query, (new_name, old_name))
            self.commit()
    
    def __insert_chall(self, category_id: int) -> int:
        challenge_query = Template(self.config.get('chall_query')).substitute({'CATEGORY_ID': category_id})
    
        self.cursor.execute(challenge_query)
        rows = self.cursor.fetchall()
        
        return rows[0][0]
    
    def __insert_hints(self, chall_id: int) -> list:
        if self.config.get("hints_changed"):
            self.__delete_hints(chall_id)
            
            hints_query = Template(self.config.get('hints_query')).substitute({'CHALL_ID': chall_id})
        
            self.cursor.execute(hints_query)
            rows = self.cursor.fetchall()
            
            hids = list(map(lambda x: x[0], rows))
        
            self.__update_hints_in_chall(chall_id, hids)
    
    def __update_hints_in_chall(self, chall_id: int, hint_ids: list) -> None:
        query = """
        UPDATE challenges
        SET hints = %s
        WHERE chall_id = %s
        """
        
        self.cursor.execute(query, (hint_ids, chall_id))
    
    def __delete_hints(self, chall_id: int) -> None:
        query = """
        SELECT hints FROM challenges
        WHERE chall_id = %s
        """
        
        self.cursor.execute(query, (chall_id,))
        rows = self.cursor.fetchall()
        
        hint_ids = rows[0][0]
        
        for hid in hint_ids:
            delete_query = f"DELETE FROM hints WHERE hid = {hid}"
            self.cursor.execute(delete_query)
    
    def update_all(self) -> int:
        self.__change_chall_name()
        
        category_id = self.__insert_category()
        chall_id = self.__insert_chall(category_id)
        hids = self.__insert_hints(chall_id)
        
        return chall_id
        
    def update_links(self, chall_id: int, links: list) -> None:
        query = """
        UPDATE challenges
        SET files = %s
        WHERE chall_id = %s
        """
        
        self.cursor.execute(query, (links, chall_id))
        
    def commit(self):
        self.conn.commit()
    
    def close(self):
        if self.cursor:
            self.cursor.close()
        if self.conn:
            self.conn.close()