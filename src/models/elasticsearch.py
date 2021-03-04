
import requests
import json


class Elasticsearch(dict):

    def __init__(self, address, verify_ssl, logger, auth=('user', 'pass')):
        self.address = address
        self.auth = auth
        self.logger = logger
        self.headers = {
            "Content-Type": "application/json"
        }
        self.verify_ssl = verify_ssl

        dict.__init__(self, address=address)

        r = requests.get("{}/{}".format(self.address, "_security/_authenticate"), auth=self.auth, verify=self.verify_ssl, headers=self.headers)

        if r.status_code != 200:
            raise Exception("Elasticsearch authentication error (management user)")

    def check_user(self, username) -> bool:
        r = requests.get("{}/{}/{}".format(self.address, "_security/user", username), auth=self.auth, verify=self.verify_ssl, headers=self.headers)

        if r.status_code == 404:
            return False

        return True

    def update_user(self, username, password, email, full_name, metadata, roles):

        request_body = {
            "enabled": True,
            "email": email,
            "password": password,
            "metadata": metadata,
            "full_name": full_name,
            "roles": roles,
        }

        r = requests.post("{}/{}/{}".format(self.address, "_security/user", username), auth=self.auth, verify=self.verify_ssl, data=json.dumps(request_body), headers=self.headers)

        self.logger.debug(r.text)

        if r.status_code != 200:
            raise Exception('Error whilst creating/updating user')

        response = r.json()
        return True if 'created' in response and response['created'] else False
