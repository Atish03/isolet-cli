import psycopg2, json
import psycopg2.extras
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
        self.cursor = self.conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)
    
    def __insert_category(self) -> int:
        catergory_values = tuple(self.config.get('category_values'))
        
        query = """
        INSERT INTO categories
        (category_name)
        VALUES (%s)
        ON CONFLICT (category_name)
        DO UPDATE SET
            category_name = EXCLUDED.category_name
        RETURNING category_id
        """
        
        self.cursor.execute(query, catergory_values)
        row = self.cursor.fetchone()
        
        return row["category_id"]
    
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
        challenge_values = tuple(self.config.get('chall_values'))
        
        query = f"""
        INSERT INTO challenges
        (chall_name, category_id, type, prompt, points, flag, author, visible, tags, links, subd, port, deployment, attempts)
        VALUES (%s, {category_id}, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (chall_name)
        DO UPDATE SET
            category_id = EXCLUDED.category_id,
            type = EXCLUDED.type,
            prompt = EXCLUDED.prompt,
            points = EXCLUDED.points,
            flag = EXCLUDED.flag,
            author = EXCLUDED.author,
            visible = EXCLUDED.visible,
            tags = EXCLUDED.tags,
            links = EXCLUDED.links,
            subd = EXCLUDED.subd,
            port = EXCLUDED.port,
            deployment = EXCLUDED.deployment,
            attempts = EXCLUDED.attempts
        RETURNING chall_id
        """
    
        self.cursor.execute(query, challenge_values)
        row = self.cursor.fetchone()
        
        return row["chall_id"]
    
    def __insert_hints(self, chall_id: int) -> list:
        if self.config.get("hints_changed"):
            self.__delete_hints(chall_id)
            
            hint_values = self.config.get("hints_values")
            
            hints_query = f"""
            INSERT INTO hints
            (chall_id, hint, cost, visible)
            VALUES ({chall_id}, %s, %s, %s)
            RETURNING hid
            """

            hids = []
            
            for hint in hint_values:
                self.cursor.execute(hints_query, tuple(hint))
                row = self.cursor.fetchone()
                hids.append(row["hid"])
        
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
        row = self.cursor.fetchone()
        
        hint_ids = row["hints"]
        
        for hid in hint_ids:
            delete_query = f"DELETE FROM hints WHERE hid = {hid}"
            self.cursor.execute(delete_query)
    
    def update_all(self) -> int:
        self.__change_chall_name()
        
        category_id = self.__insert_category()
        chall_id = self.__insert_chall(category_id)
        hids = self.__insert_hints(chall_id)
        
        return chall_id
    
    def __get_prev_files(self, chall_id) -> list:
        query = """
        SELECT files FROM challenges
        WHERE chall_id = %s
        """
        
        self.cursor.execute(query, (chall_id,))
        row = self.cursor.fetchone()
        
        return row["files"]
        
    def update_links(self, chall_id: int, links: list) -> None:
        query = """
        UPDATE challenges
        SET files = %s
        WHERE chall_id = %s
        """
        
        # prev_files = self.__get_prev_files(chall_id)
        # prev_files += links
        
        # set_files = set(prev_files)
        
        self.cursor.execute(query, (links, chall_id))
        
    def commit(self):
        self.conn.commit()
    
    def close(self):
        if self.cursor:
            self.cursor.close()
        if self.conn:
            self.conn.close()